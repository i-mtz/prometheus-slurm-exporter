/* Copyright 2020 Victor Penso

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. */

package main

import (
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func UsersData() []byte {
	cmd := exec.Command("squeue", "-a", "-r", "-h", "--noconvert", "-o %A|%u|%a|%q|%T|%C|%m|%b")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	out, _ := ioutil.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	return out
}

// gpuTresRe matches a GPU count inside a TRES/GRES specification such as
// "gres/gpu:4" or "gres/gpu:titan:4".
var gpuTresRe = regexp.MustCompile(`gres/gpu(:[^:]+)?:(\d+)`)

// parseGpuCount extracts the GPU count from a TRES/GRES string, e.g.
// "gres/gpu:4" or "gres/gpu:titan:4". It returns 0 when no GPU count is present.
func parseGpuCount(tres string) float64 {
	m := gpuTresRe.FindStringSubmatch(tres)
	if m == nil {
		return 0
	}
	n, _ := strconv.ParseFloat(m[2], 64)
	return n
}

type UserJobMetrics struct {
	user           string
	account        string
	qos            string
	pending        float64
	pending_cpus   float64
	pending_memory float64
	pending_gpus   float64
	running        float64
	running_cpus   float64
	running_memory float64
	running_gpus   float64
	suspended      float64
}

func ParseUsersMetrics(input []byte) map[string]*UserJobMetrics {
	users := make(map[string]*UserJobMetrics)
	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			user := parts[1]
			account := parts[2]
			qos := parts[3]
			key := user + "|" + account + "|" + qos
			_, ok := users[key]
			if !ok {
				users[key] = &UserJobMetrics{user, account, qos, 0, 0, 0, 0, 0, 0, 0, 0, 0}
			}
			state := parts[4]
			state = strings.ToLower(state)
			cpus, _ := strconv.ParseFloat(parts[5], 64)
			memory, _ := strconv.ParseFloat(strings.TrimSuffix(parts[6], "M"), 64)
			var gpus float64
			if len(parts) > 7 {
				gpus = parseGpuCount(parts[7])
			}
			pending := regexp.MustCompile(`^pending`)
			running := regexp.MustCompile(`^running`)
			suspended := regexp.MustCompile(`^suspended`)
			switch {
			case pending.MatchString(state) == true:
				users[key].pending++
				users[key].pending_cpus += cpus
				users[key].pending_memory += memory
				users[key].pending_gpus += gpus
			case running.MatchString(state) == true:
				users[key].running++
				users[key].running_cpus += cpus
				users[key].running_memory += memory
				users[key].running_gpus += gpus
			case suspended.MatchString(state) == true:
				users[key].suspended++
			}
		}
	}
	return users
}

type UsersCollector struct {
	pending        *prometheus.Desc
	pending_cpus   *prometheus.Desc
	pending_memory *prometheus.Desc
	pending_gpus   *prometheus.Desc
	running        *prometheus.Desc
	running_cpus   *prometheus.Desc
	running_memory *prometheus.Desc
	running_gpus   *prometheus.Desc
	suspended      *prometheus.Desc
}

func NewUsersCollector() *UsersCollector {
	labels := []string{"user", "account", "qos"}
	return &UsersCollector{
		pending:        prometheus.NewDesc("slurm_user_jobs_pending", "Pending jobs for user", labels, nil),
		pending_cpus:   prometheus.NewDesc("slurm_user_cpus_pending", "Pending jobs for user", labels, nil),
		pending_memory: prometheus.NewDesc("slurm_user_memory_pending", "Pending jobs for user", labels, nil),
		pending_gpus:   prometheus.NewDesc("slurm_user_gpus_pending", "Pending GPUs for user", labels, nil),
		running:        prometheus.NewDesc("slurm_user_jobs_running", "Running jobs for user", labels, nil),
		running_cpus:   prometheus.NewDesc("slurm_user_cpus_running", "Running cpus for user", labels, nil),
		running_memory: prometheus.NewDesc("slurm_user_memory_running", "Running memory for user", labels, nil),
		running_gpus:   prometheus.NewDesc("slurm_user_gpus_running", "Running GPUs for user", labels, nil),
		suspended:      prometheus.NewDesc("slurm_user_jobs_suspended", "Suspended jobs for user", labels, nil),
	}
}

func (uc *UsersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- uc.pending
	ch <- uc.pending_cpus
	ch <- uc.pending_memory
	ch <- uc.pending_gpus
	ch <- uc.running
	ch <- uc.running_cpus
	ch <- uc.running_memory
	ch <- uc.running_gpus
	ch <- uc.suspended
}

func (uc *UsersCollector) Collect(ch chan<- prometheus.Metric) {
	um := ParseUsersMetrics(UsersData())
	for u := range um {
		if um[u].pending > 0 {
			ch <- prometheus.MustNewConstMetric(uc.pending, prometheus.GaugeValue, um[u].pending, um[u].user, um[u].account, um[u].qos)
		}
		if um[u].pending_cpus > 0 {
			ch <- prometheus.MustNewConstMetric(uc.pending_cpus, prometheus.GaugeValue, um[u].pending_cpus, um[u].user, um[u].account, um[u].qos)
		}
		if um[u].pending_memory > 0 {
			ch <- prometheus.MustNewConstMetric(uc.pending_memory, prometheus.GaugeValue, um[u].pending_memory, um[u].user, um[u].account, um[u].qos)
		}
		if um[u].pending_gpus > 0 {
			ch <- prometheus.MustNewConstMetric(uc.pending_gpus, prometheus.GaugeValue, um[u].pending_gpus, um[u].user, um[u].account, um[u].qos)
		}
		if um[u].running > 0 {
			ch <- prometheus.MustNewConstMetric(uc.running, prometheus.GaugeValue, um[u].running, um[u].user, um[u].account, um[u].qos)
		}
		if um[u].running_cpus > 0 {
			ch <- prometheus.MustNewConstMetric(uc.running_cpus, prometheus.GaugeValue, um[u].running_cpus, um[u].user, um[u].account, um[u].qos)
		}
		if um[u].running_memory > 0 {
			ch <- prometheus.MustNewConstMetric(uc.running_memory, prometheus.GaugeValue, um[u].running_memory, um[u].user, um[u].account, um[u].qos)
		}
		if um[u].running_gpus > 0 {
			ch <- prometheus.MustNewConstMetric(uc.running_gpus, prometheus.GaugeValue, um[u].running_gpus, um[u].user, um[u].account, um[u].qos)
		}
		if um[u].suspended > 0 {
			ch <- prometheus.MustNewConstMetric(uc.suspended, prometheus.GaugeValue, um[u].suspended, um[u].user, um[u].account, um[u].qos)
		}
	}
}
