/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package goproxmox_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	goproxmox "github.com/sergelogvinov/go-proxmox"

	"k8s.io/utils/ptr"
)

func TestVMCloudInitIPConfig_UnmarshalString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		ipconfig goproxmox.VMCloudInitIPConfig
	}{
		{
			name:     "empty",
			template: "",
			ipconfig: goproxmox.VMCloudInitIPConfig{},
		},
		{
			name:     "ipv4-only",
			template: "ip=1.2.3.4,gw=1.2.3.1",
			ipconfig: goproxmox.VMCloudInitIPConfig{
				GatewayIPv4: "1.2.3.1",
				IPv4:        "1.2.3.4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := goproxmox.VMCloudInitIPConfig{}

			err := res.UnmarshalString(tt.template)
			assert.NoError(t, err)
			assert.Equal(t, tt.ipconfig, res)
		})
	}
}

func TestVMNetworkDevice_UnmarshalString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		iface    goproxmox.VMNetworkDevice
	}{
		{
			name:     "empty",
			template: "",
			iface:    goproxmox.VMNetworkDevice{},
		},
		{
			name:     "virtio",
			template: "virtio=32:90:AC:10:00:91,bridge=vmbr0,firewall=1,mtu=1500,queues=8",
			iface: goproxmox.VMNetworkDevice{
				Virtio:   "32:90:AC:10:00:91",
				Bridge:   "vmbr0",
				Firewall: goproxmox.NewIntOrBool(true),
				MTU:      ptr.To(1500),
				Queues:   ptr.To(8),
			},
		},
		{
			name:     "virtio",
			template: "virtio=32:90:AC:10:00:91,bridge=vmbr0,firewall=1,mtu=1500,queues=8,tag=1,trunks=1;2",
			iface: goproxmox.VMNetworkDevice{
				Virtio:   "32:90:AC:10:00:91",
				Bridge:   "vmbr0",
				Firewall: goproxmox.NewIntOrBool(true),
				MTU:      ptr.To(1500),
				Queues:   ptr.To(8),
				Tag:      ptr.To(1),
				Trunks:   []int{1, 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := goproxmox.VMNetworkDevice{}

			err := res.UnmarshalString(tt.template)
			assert.NoError(t, err)
			assert.Equal(t, tt.iface, res)
		})
	}
}

func TestVMNetworkDevice_ToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		iface goproxmox.VMNetworkDevice
		res   string
	}{
		{
			name:  "empty",
			iface: goproxmox.VMNetworkDevice{},
			res:   "",
		},
		{
			name: "virtio",
			iface: goproxmox.VMNetworkDevice{
				Virtio:   "32:90:AC:10:00:91",
				Bridge:   "vmbr0",
				Firewall: goproxmox.NewIntOrBool(true),
				MTU:      ptr.To(1500),
				Queues:   ptr.To(8),
				Trunks:   []int{1, 2},
			},
			res: "virtio=32:90:AC:10:00:91,bridge=vmbr0,firewall=1,mtu=1500,queues=8,trunks=1;2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := tt.iface.ToString()

			assert.NoError(t, err)
			assert.Equal(t, tt.res, res)
		})
	}
}

func TestVMNUMA_UnmarshalString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		numa     goproxmox.VMNUMA
	}{
		{
			name:     "empty",
			template: "",
			numa:     goproxmox.VMNUMA{},
		},
		{
			name:     "numa0",
			template: "cpus=0-3,hostnodes=0,memory=12288,policy=bind",
			numa: goproxmox.VMNUMA{
				CPUIDs:        []string{"0-3"},
				HostNodeNames: []string{"0"},
				Memory:        ptr.To(12288),
				Policy:        "bind",
			},
		},
		{
			name:     "numa1",
			template: "cpus=4-7,hostnodes=1,memory=12288",
			numa: goproxmox.VMNUMA{
				CPUIDs:        []string{"4-7"},
				HostNodeNames: []string{"1"},
				Memory:        ptr.To(12288),
				Policy:        "",
			},
		},
		{
			name:     "numa2",
			template: "cpus=0-3;4-7,hostnodes=0;1,memory=12288",
			numa: goproxmox.VMNUMA{
				CPUIDs:        []string{"0-3", "4-7"},
				HostNodeNames: []string{"0", "1"},
				Memory:        ptr.To(12288),
				Policy:        "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := goproxmox.VMNUMA{}

			err := res.UnmarshalString(tt.template)
			assert.NoError(t, err)
			assert.Equal(t, tt.numa, res)
		})
	}
}

func TestVMNUMA_ToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		numa goproxmox.VMNUMA
		res  string
	}{
		{
			name: "empty",
			numa: goproxmox.VMNUMA{},
			res:  "",
		},
		{
			name: "numa0",
			numa: goproxmox.VMNUMA{
				CPUIDs:        []string{"0-3"},
				HostNodeNames: []string{"0"},
				Memory:        ptr.To(12288),
				Policy:        "bind",
			},
			res: "cpus=0-3,hostnodes=0,memory=12288,policy=bind",
		},
		{
			name: "numa1",
			numa: goproxmox.VMNUMA{
				CPUIDs:        []string{"4-7"},
				HostNodeNames: []string{"1"},
				Memory:        ptr.To(12288),
			},
			res: "cpus=4-7,hostnodes=1,memory=12288",
		},
		{
			name: "numa2",
			numa: goproxmox.VMNUMA{
				CPUIDs:        []string{"0-3", "4-7"},
				HostNodeNames: []string{"0", "1"},
				Memory:        ptr.To(12288),
			},
			res: "cpus=0-3;4-7,hostnodes=0;1,memory=12288",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := tt.numa.ToString()

			assert.NoError(t, err)
			assert.Equal(t, tt.res, res)
		})
	}
}
