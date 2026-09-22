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

	assert.Equal(t, 9, len(metrics))

	// user1/group1/a100-1week: 1 PENDING job, 8 cpus, 65536M
	m := metrics["user1|group1|a100-1week"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(1), m.pending)
	assert.Equal(t, float64(8), m.pending_cpus)
	assert.Equal(t, float64(65536), m.pending_memory)

	// user2/group2/a100-1day: 16 PENDING jobs, 192 cpus, 1392640M
	m = metrics["user2|group2|a100-1day"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(16), m.pending)
	assert.Equal(t, float64(192), m.pending_cpus)
	assert.Equal(t, float64(1392640), m.pending_memory)

	// user3/group3/fast: 38 RUNNING jobs, 152 cpus, 311296M
	m = metrics["user3|group3|fast"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(38), m.running)
	assert.Equal(t, float64(152), m.running_cpus)
	assert.Equal(t, float64(311296), m.running_memory)

	// user7/group7 is split across two qos values
	m = metrics["user7|group7|titan-6hours"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(9), m.running)
	assert.Equal(t, float64(18), m.running_cpus)
	assert.Equal(t, float64(442368), m.running_memory)

	m = metrics["user7|group7|titan-1day"]
	assert.NotNil(t, m)
	assert.Equal(t, float64(15), m.running)
	assert.Equal(t, float64(30), m.running_cpus)
	assert.Equal(t, float64(737280), m.running_memory)

	// Verify the stored dimensions match the key
	for key, m := range metrics {
		parts := strings.SplitN(key, "|", 3)
		assert.Equal(t, parts[0], m.user)
		assert.Equal(t, parts[1], m.account)
		assert.Equal(t, parts[2], m.qos)
	}
}
