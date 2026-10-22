// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

// Handler for "raw" containers.
package raw

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/cadvisor/lib/container"
	"github.com/google/cadvisor/lib/container/common"
	"github.com/google/cadvisor/lib/fs"
	info "github.com/google/cadvisor/lib/model"
)

func TestFsToFsStats(t *testing.T) {
	inodes := uint64(100)
	inodesFree := uint64(50)
	testCases := map[string]struct {
		fs       *fs.Fs
		expected info.FsStats
	}{
		"has_inodes": {
			fs: &fs.Fs{
				DeviceInfo: fs.DeviceInfo{Device: "123"},
				Type:       fs.VFS,
				Capacity:   uint64(1024 * 1024),
				Free:       uint64(1024),
				Available:  uint64(1024),
				Inodes:     &inodes,
				InodesFree: &inodesFree,
				DiskStats: fs.DiskStats{
					ReadsCompleted:  uint64(100),
					ReadsMerged:     uint64(100),
					SectorsRead:     uint64(100),
					ReadTime:        uint64(100),
					WritesCompleted: uint64(100),
					WritesMerged:    uint64(100),
					SectorsWritten:  uint64(100),
					WriteTime:       uint64(100),
					IoInProgress:    uint64(100),
					IoTime:          uint64(100),
					WeightedIoTime:  uint64(100),
				},
			},
			expected: info.FsStats{
				Device:          "123",
				Type:            fs.VFS.String(),
				Limit:           uint64(1024 * 1024),
				Usage:           uint64(1024*1024) - uint64(1024),
				HasInodes:       true,
				Inodes:          inodes,
				InodesFree:      inodesFree,
				Available:       uint64(1024),
				ReadsCompleted:  uint64(100),
				ReadsMerged:     uint64(100),
				SectorsRead:     uint64(100),
				ReadTime:        uint64(100),
				WritesCompleted: uint64(100),
				WritesMerged:    uint64(100),
				SectorsWritten:  uint64(100),
				WriteTime:       uint64(100),
				IoInProgress:    uint64(100),
				IoTime:          uint64(100),
				WeightedIoTime:  uint64(100),
			},
		},
		"has_no_inodes": {
			fs: &fs.Fs{
				DeviceInfo: fs.DeviceInfo{Device: "123"},
				Type:       fs.DeviceMapper,
				Capacity:   uint64(1024 * 1024),
				Free:       uint64(1024),
				Available:  uint64(1024),
				DiskStats: fs.DiskStats{
					ReadsCompleted:  uint64(100),
					ReadsMerged:     uint64(100),
					SectorsRead:     uint64(100),
					ReadTime:        uint64(100),
					WritesCompleted: uint64(100),
					WritesMerged:    uint64(100),
					SectorsWritten:  uint64(100),
					WriteTime:       uint64(100),
					IoInProgress:    uint64(100),
					IoTime:          uint64(100),
					WeightedIoTime:  uint64(100),
				},
			},
			expected: info.FsStats{
				Device:          "123",
				Type:            fs.DeviceMapper.String(),
				Limit:           uint64(1024 * 1024),
				Usage:           uint64(1024*1024) - uint64(1024),
				HasInodes:       false,
				Available:       uint64(1024),
				ReadsCompleted:  uint64(100),
				ReadsMerged:     uint64(100),
				SectorsRead:     uint64(100),
				ReadTime:        uint64(100),
				WritesCompleted: uint64(100),
				WritesMerged:    uint64(100),
				SectorsWritten:  uint64(100),
				WriteTime:       uint64(100),
				IoInProgress:    uint64(100),
				IoTime:          uint64(100),
				WeightedIoTime:  uint64(100),
			},
		},
	}
	for testName, testCase := range testCases {
		actual := fsToFsStats(testCase.fs)
		if !reflect.DeepEqual(testCase.expected, actual) {
			t.Errorf("test case=%v, expected=%v, actual=%v", testName, testCase.expected, actual)
		}
	}
}

func TestGetFsStats(t *testing.T) {
	inodes := uint64(2000)
	inodesFree := uint64(1000)

	cases := map[string]struct {
		name                string
		includedMetrics     container.MetricSet
		externalMounts      []common.Mount
		globalFsInfo        func() ([]fs.Fs, error)
		getFsInfoForPath    func(mountSet map[string]struct{}) ([]fs.Fs, error)
		diskIO              info.DiskIoStats
		expectedFilesystems []info.FsStats
		expectedDiskIO      info.DiskIoStats
	}{
		"root with disk metrics enabled": {
			name:            "/",
			includedMetrics: container.MetricSet{container.DiskUsageMetrics: struct{}{}, container.DiskIOMetrics: struct{}{}},
			externalMounts:  []common.Mount{},
			globalFsInfo: func() ([]fs.Fs, error) {
				return []fs.Fs{{
					DeviceInfo: fs.DeviceInfo{
						Device: "123",
						Major:  1,
						Minor:  2,
					},
					Type:       "devicemapper",
					Capacity:   1000,
					Free:       500,
					Available:  450,
					Inodes:     &inodes,
					InodesFree: &inodesFree,
					DiskStats: fs.DiskStats{
						ReadsCompleted:  1,
						ReadsMerged:     2,
						SectorsRead:     3,
						ReadTime:        3,
						WritesCompleted: 4,
						WritesMerged:    6,
						SectorsWritten:  7,
						WriteTime:       8,
						IoInProgress:    9,
						IoTime:          10,
						WeightedIoTime:  11,
					},
				}}, nil
			},
			expectedFilesystems: []info.FsStats{
				{
					Device:          "123",
					Type:            "devicemapper",
					Limit:           1000,
					Usage:           500,
					BaseUsage:       0,
					Available:       450,
					HasInodes:       true,
					Inodes:          2000,
					InodesFree:      1000,
					ReadsCompleted:  1,
					ReadsMerged:     2,
					SectorsRead:     3,
					ReadTime:        3,
					WritesCompleted: 4,
					WritesMerged:    6,
					SectorsWritten:  7,
					WriteTime:       8,
					IoInProgress:    9,
					IoTime:          10,
					WeightedIoTime:  11,
				},
			},
			expectedDiskIO: info.DiskIoStats{
				IoServiceBytes: []info.PerDiskStats{
					{
						Device: "123",
						Major:  1,
						Minor:  2,
						Stats:  map[string]uint64{"a": 1},
					},
				},
			},
			diskIO: info.DiskIoStats{
				IoServiceBytes: []info.PerDiskStats{
					{
						Device: "",
						Major:  1,
						Minor:  2,
						Stats:  map[string]uint64{"a": 1},
					},
				},
			},
		},
		"root with disk usage metrics enabled": {
			name:            "/",
			includedMetrics: container.MetricSet{container.DiskUsageMetrics: struct{}{}},
			externalMounts:  []common.Mount{},
			globalFsInfo: func() ([]fs.Fs, error) {
				return []fs.Fs{{
					DeviceInfo: fs.DeviceInfo{
						Device: "123",
						Major:  1,
						Minor:  2,
					},
					Type:       "devicemapper",
					Capacity:   1000,
					Free:       500,
					Available:  450,
					Inodes:     &inodes,
					InodesFree: &inodesFree,
					DiskStats: fs.DiskStats{
						ReadsCompleted:  1,
						ReadsMerged:     2,
						SectorsRead:     3,
						ReadTime:        3,
						WritesCompleted: 4,
						WritesMerged:    6,
						SectorsWritten:  7,
						WriteTime:       8,
						IoInProgress:    9,
						IoTime:          10,
						WeightedIoTime:  11,
					},
				}}, nil
			},
			expectedFilesystems: []info.FsStats{
				{
					Device:          "123",
					Type:            "devicemapper",
					Limit:           1000,
					Usage:           500,
					BaseUsage:       0,
					Available:       450,
					HasInodes:       true,
					Inodes:          2000,
					InodesFree:      1000,
					ReadsCompleted:  1,
					ReadsMerged:     2,
					SectorsRead:     3,
					ReadTime:        3,
					WritesCompleted: 4,
					WritesMerged:    6,
					SectorsWritten:  7,
					WriteTime:       8,
					IoInProgress:    9,
					IoTime:          10,
					WeightedIoTime:  11,
				},
			},
			// DiskIoStats must not be enriched with device name.
			// This is imperfect check but I can't find any other.
			expectedDiskIO: info.DiskIoStats{
				IoServiceBytes: []info.PerDiskStats{
					{
						Device: "",
						Major:  1,
						Minor:  2,
						Stats:  map[string]uint64{"a": 1},
					},
				},
			},
			diskIO: info.DiskIoStats{
				IoServiceBytes: []info.PerDiskStats{
					{
						Device: "",
						Major:  1,
						Minor:  2,
						Stats:  map[string]uint64{"a": 1},
					},
				},
			},
		},
		"root with disk I/O metrics enabled": {
			name:            "/",
			includedMetrics: container.MetricSet{container.DiskIOMetrics: struct{}{}},
			externalMounts:  []common.Mount{},
			globalFsInfo: func() ([]fs.Fs, error) {
				return []fs.Fs{{
					DeviceInfo: fs.DeviceInfo{
						Device: "123",
						Major:  1,
						Minor:  2,
					},
					Type:       "devicemapper",
					Capacity:   1000,
					Free:       500,
					Available:  450,
					Inodes:     &inodes,
					InodesFree: &inodesFree,
					DiskStats: fs.DiskStats{
						ReadsCompleted:  1,
						ReadsMerged:     2,
						SectorsRead:     3,
						ReadTime:        3,
						WritesCompleted: 4,
						WritesMerged:    6,
						SectorsWritten:  7,
						WriteTime:       8,
						IoInProgress:    9,
						IoTime:          10,
						WeightedIoTime:  11,
					},
				}}, nil
			},
			expectedDiskIO: info.DiskIoStats{
				IoServiceBytes: []info.PerDiskStats{
					{
						Device: "123",
						Major:  1,
						Minor:  2,
						Stats:  map[string]uint64{"a": 1},
					},
				},
			},
			diskIO: info.DiskIoStats{
				IoServiceBytes: []info.PerDiskStats{
					{
						Device: "",
						Major:  1,
						Minor:  2,
						Stats:  map[string]uint64{"a": 1},
					},
				},
			},
		},
		"root with disk metrics disabled": {
			name:            "/",
			includedMetrics: container.MetricSet{},
			externalMounts:  []common.Mount{},
			// DiskIoStats must not be enriched with device name.
			// This is imperfect check but I can't find any other.
			expectedDiskIO: info.DiskIoStats{
				IoServiceBytes: []info.PerDiskStats{
					{
						Device: "",
						Major:  1,
						Minor:  2,
						Stats:  map[string]uint64{"a": 1},
					},
				},
			},
			diskIO: info.DiskIoStats{
				IoServiceBytes: []info.PerDiskStats{
					{
						Device: "",
						Major:  1,
						Minor:  2,
						Stats:  map[string]uint64{"a": 1},
					},
				},
			},
		},
		"random container with disk metrics enabled": {
			name:            "/random/container",
			includedMetrics: container.MetricSet{container.DiskUsageMetrics: struct{}{}, container.DiskIOMetrics: struct{}{}},
			externalMounts: []common.Mount{
				{HostDir: "/", ContainerDir: "/"},
				{HostDir: "/random", ContainerDir: "/completely/random"},
			},
			getFsInfoForPath: func(mountSet map[string]struct{}) ([]fs.Fs, error) {
				return []fs.Fs{
					{
						DeviceInfo: fs.DeviceInfo{
							Device: "123",
							Major:  1,
							Minor:  2,
						},
						Type:       "devicemapper",
						Capacity:   1000,
						Free:       500,
						Available:  450,
						Inodes:     &inodes,
						InodesFree: &inodesFree,
						DiskStats: fs.DiskStats{
							ReadsCompleted:  1,
							ReadsMerged:     2,
							SectorsRead:     3,
							ReadTime:        3,
							WritesCompleted: 4,
							WritesMerged:    6,
							SectorsWritten:  7,
							WriteTime:       8,
							IoInProgress:    9,
							IoTime:          10,
							WeightedIoTime:  11,
						},
					},
					{
						DeviceInfo: fs.DeviceInfo{
							Device: "246",
							Major:  3,
							Minor:  4,
						},
						Type:       "devicemapper",
						Capacity:   2000,
						Free:       1000,
						Available:  900,
						Inodes:     &inodes,
						InodesFree: &inodesFree,
						DiskStats: fs.DiskStats{
							ReadsCompleted:  10,
							ReadsMerged:     20,
							SectorsRead:     25,
							ReadTime:        30,
							WritesCompleted: 40,
							WritesMerged:    60,
							SectorsWritten:  70,
							WriteTime:       80,
							IoInProgress:    90,
							IoTime:          100,
							WeightedIoTime:  110,
						},
					},
				}, nil
			},
			expectedFilesystems: []info.FsStats{
				{
					Device:          "123",
					Type:            "devicemapper",
					Limit:           1000,
					Usage:           500,
					BaseUsage:       0,
					Available:       450,
					HasInodes:       true,
					Inodes:          2000,
					InodesFree:      1000,
					ReadsCompleted:  1,
					ReadsMerged:     2,
					SectorsRead:     3,
					ReadTime:        3,
					WritesCompleted: 4,
					WritesMerged:    6,
					SectorsWritten:  7,
					WriteTime:       8,
					IoInProgress:    9,
					IoTime:          10,
					WeightedIoTime:  11,
				},
				{
					Device:          "246",
					Type:            "devicemapper",
					Limit:           2000,
					Usage:           1000,
					BaseUsage:       0,
					Available:       900,
					HasInodes:       true,
					Inodes:          2000,
					InodesFree:      1000,
					ReadsCompleted:  10,
					ReadsMerged:     20,
					SectorsRead:     25,
					ReadTime:        30,
					WritesCompleted: 40,
					WritesMerged:    60,
					SectorsWritten:  70,
					WriteTime:       80,
					IoInProgress:    90,
					IoTime:          100,
					WeightedIoTime:  110,
				},
			},
			expectedDiskIO: info.DiskIoStats{
				IoServiceBytes: []info.PerDiskStats{
					{
						Device: "123",
						Major:  1,
						Minor:  2,
						Stats:  map[string]uint64{"a": 1},
					},
					{
						Device: "246",
						Major:  3,
						Minor:  4,
						Stats:  map[string]uint64{"b": 2},
					},
				},
			},
			diskIO: info.DiskIoStats{
				IoServiceBytes: []info.PerDiskStats{
					{
						Device: "",
						Major:  1,
						Minor:  2,
						Stats:  map[string]uint64{"a": 1},
					},
					{
						Device: "",
						Major:  3,
						Minor:  4,
						Stats:  map[string]uint64{"b": 2},
					},
				},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			handler := rawContainerHandler{
				name:               c.name,
				includedMetrics:    c.includedMetrics,
				fsInfo:             fsInfo{c.globalFsInfo, c.getFsInfoForPath},
				externalMounts:     c.externalMounts,
				machineInfoFactory: machineInfo{},
			}
			stats := &info.ContainerStats{DiskIo: &c.diskIO}
			err := handler.getFsStats(stats)

			assert.NoError(t, err)
			assert.Len(t, stats.Filesystem, len(c.expectedFilesystems))
			assert.Equal(t, c.expectedFilesystems, stats.Filesystem)
			assert.Equal(t, c.expectedDiskIO, *stats.DiskIo)
		})
	}
}

type fsInfo struct {
	globalFsInfo     func() ([]fs.Fs, error)
	getFsInfoForPath func(mountSet map[string]struct{}) ([]fs.Fs, error)
}

func (f fsInfo) GetGlobalFsInfo() ([]fs.Fs, error) {
	return f.globalFsInfo()
}

func (f fsInfo) GetFsInfoForPath(mountSet map[string]struct{}) ([]fs.Fs, error) {
	return f.getFsInfoForPath(mountSet)
}

func (f fsInfo) GetDirUsage(_ string) (fs.UsageInfo, error) {
	panic("unsupported")
}

func (f fsInfo) GetDeviceInfoByFsUUID(_ string) (*fs.DeviceInfo, error) {
	panic("unsupported")
}

func (f fsInfo) GetDirFsDevice(_ string) (*fs.DeviceInfo, error) {
	panic("unsupported")
}

func (f fsInfo) GetDeviceForLabel(_ string) (string, error) {
	panic("unsupported")
}

func (f fsInfo) GetLabelsForDevice(_ string) ([]string, error) {
	panic("unsupported")
}

func (f fsInfo) GetMountpointForDevice(_ string) (string, error) {
	panic("unsupported")
}

type machineInfo struct{}

func (m machineInfo) GetMachineInfo() (*info.MachineInfo, error) {
	return &info.MachineInfo{
		DiskMap: map[string]info.DiskInfo{
			"1:2": {
				Name:      "sda",
				Size:      1234,
				Scheduler: "none",
			},
			"3:4": {
				Name:      "sdb",
				Size:      5678,
				Scheduler: "none",
			},
		},
	}, nil
}

func (m machineInfo) GetVersionInfo() (*info.VersionInfo, error) {
	panic("unsupported")
}

const testProcNetDev = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo:       1       1    0    0    0     0          0         0        1       1    0    0    0     0       0          0
  eth0:     100       2    0    0    0     0          0         0      200       3    0    0    0     0       0          0
`

// makeNetDev creates rootFs/proc/<pid>/net/dev, as a hard link to the one of
// sameNetns if it is not 0.
func makeNetDev(t *testing.T, rootFs string, pid, sameNetns int) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(filepath.Join(rootFs, "proc", strconv.Itoa(pid), "net"), 0o755))
	if sameNetns != 0 {
		assert.NoError(t, os.Link(netDevPath(rootFs, sameNetns), netDevPath(rootFs, pid)))
		return
	}
	assert.NoError(t, os.WriteFile(netDevPath(rootFs, pid), []byte(testProcNetDev), 0o644))
}

func TestFindNetnsPid(t *testing.T) {
	rootFs := t.TempDir()
	makeNetDev(t, rootFs, 1, 0)
	makeNetDev(t, rootFs, 5, 1)
	makeNetDev(t, rootFs, 20, 0)
	makeNetDev(t, rootFs, 30, 20)
	hostNetDev, err := os.Stat(netDevPath(rootFs, 1))
	assert.NoError(t, err)

	cases := map[string]struct {
		pids        []int
		expectedPid int
	}{
		"no pids":             {pids: nil, expectedPid: 0},
		"only host pids":      {pids: []int{1, 5}, expectedPid: 0},
		"hidden pids":         {pids: []int{0, -1}, expectedPid: 0},
		"vanished pid":        {pids: []int{25, 30}, expectedPid: 30},
		"host pid first":      {pids: []int{5, 30}, expectedPid: 30},
		"lowest non-host pid": {pids: []int{30, 5, 20}, expectedPid: 20},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			pid, fi := findNetnsPid(rootFs, tc.pids, hostNetDev)
			assert.Equal(t, tc.expectedPid, pid)
			assert.Equal(t, tc.expectedPid != 0, fi != nil)
		})
	}
}

func TestGetNetworkStats(t *testing.T) {
	rootFs := t.TempDir()
	makeNetDev(t, rootFs, 1, 0)
	makeNetDev(t, rootFs, 5, 1)
	makeNetDev(t, rootFs, 20, 0)
	makeNetDev(t, rootFs, 30, 0)

	h := &rawContainerHandler{
		name:            "/system.slice/test.service/libpod-payload-0",
		rootFs:          rootFs,
		networkStats:    true,
		includedMetrics: container.MetricSet{container.NetworkUsageMetrics: struct{}{}},
	}
	assert.True(t, h.hasNetworkStats())

	stats, err := h.getNetworkStats([]int{5})
	assert.NoError(t, err)
	assert.Nil(t, stats)

	stats, err = h.getNetworkStats([]int{5, 30})
	assert.NoError(t, err)
	assert.Equal(t, 30, h.netPid)
	assert.Equal(t, []info.InterfaceStats{{Name: "eth0", RxBytes: 100, RxPackets: 2, TxBytes: 200, TxPackets: 3}}, stats.Interfaces)

	// The chosen pid is kept while it stays in the cgroup and netns.
	_, err = h.getNetworkStats([]int{20, 30})
	assert.NoError(t, err)
	assert.Equal(t, 30, h.netPid)

	// A pid that left the cgroup is replaced.
	_, err = h.getNetworkStats([]int{20})
	assert.NoError(t, err)
	assert.Equal(t, 20, h.netPid)

	// A pid that was reused in another netns is replaced.
	assert.NoError(t, os.Remove(netDevPath(rootFs, 20)))
	assert.NoError(t, os.Link(netDevPath(rootFs, 1), netDevPath(rootFs, 20)))
	stats, err = h.getNetworkStats([]int{20, 30})
	assert.NoError(t, err)
	assert.Equal(t, 30, h.netPid)
	assert.NotNil(t, stats)

	// No stats once only host netns pids are left.
	assert.NoError(t, os.RemoveAll(filepath.Join(rootFs, "proc", "30")))
	stats, err = h.getNetworkStats([]int{20, 30})
	assert.NoError(t, err)
	assert.Equal(t, 0, h.netPid)
	assert.Nil(t, stats)

	h.networkStats = false
	assert.False(t, h.hasNetworkStats())
	h.networkStats = true
	h.includedMetrics = container.MetricSet{}
	assert.False(t, h.hasNetworkStats())
	h.includedMetrics = container.MetricSet{container.NetworkUsageMetrics: struct{}{}}
	h.name = "/"
	assert.False(t, h.hasNetworkStats())
}

func TestInHostPidNamespace(t *testing.T) {
	rootFs := t.TempDir()
	assert.False(t, inHostPidNamespace(rootFs))

	assert.NoError(t, os.MkdirAll(filepath.Join(rootFs, "proc"), 0o755))
	assert.NoError(t, os.Symlink("1", filepath.Join(rootFs, "proc", "self")))
	assert.False(t, inHostPidNamespace(rootFs))

	assert.NoError(t, os.Remove(filepath.Join(rootFs, "proc", "self")))
	assert.NoError(t, os.Symlink(strconv.Itoa(os.Getpid()), filepath.Join(rootFs, "proc", "self")))
	assert.True(t, inHostPidNamespace(rootFs))
}
