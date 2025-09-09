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

package goproxmox

import (
	"github.com/luthermonson/go-proxmox"
	"k8s.io/utils/ptr"
)

func applyInstanceOptimization(vm *proxmox.VirtualMachine, options VMCloneRequest, vmOptions []proxmox.VirtualMachineOption) []proxmox.VirtualMachineOption {
	if vm.VirtualMachineConfig != nil {
		nets := vm.VirtualMachineConfig.MergeNets()

		for d, net := range nets {
			iface := VMNetworkDevice{}
			if err := iface.UnmarshalString(net); err != nil {
				return nil
			}

			iface.Queues = ptr.To(options.CPU)

			v, err := iface.ToString()
			if err != nil {
				return nil
			}

			vmOptions = append(vmOptions, proxmox.VirtualMachineOption{
				Name:  d,
				Value: v,
			})
		}
	}

	return vmOptions
}
