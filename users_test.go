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
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUsersMetrics(t *testing.T) {
	// Read the input data from a file
	file, err := os.Open("test_data/squeue_users.txt")
	if err != nil {
		t.Fatalf("Can not open test data: %v", err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatalf("Can not read test data: %v", err)
	}
	metrics := ParseUsersMetrics(data)
	t.Logf("%+v", metrics)

	assert.Equal(t, 7, len(metrics))

	// user3/account3/titan-6hours: 2 PENDING jobs, 8 cpus, 49152M, 2 gpus
	m := metrics["user3|account3|titan-6hours"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(2), m.pending)
	assert.Equal(t, float64(8), m.pending_cpus)
	assert.Equal(t, float64(49152), m.pending_memory)
	assert.Equal(t, float64(2), m.pending_gpus)

	// user3/account3/titan-1day: 2 RUNNING jobs, 8 cpus, 49152M, 2 gpus
	m = metrics["user3|account3|titan-1day"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(2), m.running)
	assert.Equal(t, float64(8), m.running_cpus)
	assert.Equal(t, float64(49152), m.running_memory)
	assert.Equal(t, float64(2), m.running_gpus)

	// user2/account2/2weeks: 5 PENDING jobs, 640 cpus, 20480M, no gpus (N/A)
	m = metrics["user2|account2|2weeks"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(5), m.pending)
	assert.Equal(t, float64(640), m.pending_cpus)
	assert.Equal(t, float64(20480), m.pending_memory)
	assert.Equal(t, float64(0), m.pending_gpus)

	// user5/account5/titan-1day: 7 RUNNING jobs, 14 cpus, 344064M, 7 gpus
	m = metrics["user5|account5|titan-1day"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(7), m.running)
	assert.Equal(t, float64(14), m.running_cpus)
	assert.Equal(t, float64(344064), m.running_memory)
	assert.Equal(t, float64(7), m.running_gpus)

	// user5/account5/titan-6hours: 17 RUNNING jobs, 34 cpus, 835584M, 17 gpus
	m = metrics["user5|account5|titan-6hours"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(17), m.running)
	assert.Equal(t, float64(34), m.running_cpus)
	assert.Equal(t, float64(835584), m.running_memory)
	assert.Equal(t, float64(17), m.running_gpus)

	// user1/account1/1week: 1 RUNNING job, 16 cpus, 40960M, no gpus (N/A)
	m = metrics["user1|account1|1week"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(1), m.running)
	assert.Equal(t, float64(16), m.running_cpus)
	assert.Equal(t, float64(40960), m.running_memory)
	assert.Equal(t, float64(0), m.running_gpus)

	// user4/account4/titan-6hours: 1 RUNNING job, 4 cpus, 61440M, 1 gpu
	m = metrics["user4|account4|titan-6hours"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(1), m.running)
	assert.Equal(t, float64(4), m.running_cpus)
	assert.Equal(t, float64(61440), m.running_memory)
	assert.Equal(t, float64(1), m.running_gpus)

	// Verify the stored dimensions match the key
	for key, m := range metrics {
		parts := strings.SplitN(key, "|", 3)
		assert.Equal(t, parts[0], m.user)
		assert.Equal(t, parts[1], m.account)
		assert.Equal(t, parts[2], m.qos)
	}
}

func TestParseGpuCount(t *testing.T) {
	assert.Equal(t, float64(0), parseGpuCount("N/A"))
	assert.Equal(t, float64(4), parseGpuCount("gres/gpu:4"))
	assert.Equal(t, float64(1), parseGpuCount("gres/gpu:titan:1"))
	assert.Equal(t, float64(16), parseGpuCount("gres/gpu:titan:16"))
	assert.Equal(t, float64(0), parseGpuCount("cpu:8+mem:1024M"))
}
