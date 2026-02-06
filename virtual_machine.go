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
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/luthermonson/go-proxmox"
)

// GetVMByID returns a VM cluster resource by its ID.
func (c *APIClient) GetVMByID(ctx context.Context, vmID uint64) (*proxmox.ClusterResource, error) {
	vmr, err := c.GetVMByFilter(ctx, func(r *proxmox.ClusterResource) (bool, error) {
		return r.VMID == vmID, nil
	})
	if err != nil {
		return nil, err
	}

	if vmr.VMID != 0 {
		return vmr, nil
	}

	return nil, ErrVirtualMachineNotFound
}

// GetVMByFilter returns a VM cluster resource by applying the provided filter functions.
func (c *APIClient) GetVMByFilter(ctx context.Context, filter ...func(*proxmox.ClusterResource) (bool, error)) (*proxmox.ClusterResource, error) {
	vmr, err := c.getResources(ctx, "vm")
	if err != nil {
		return nil, err
	}

	for _, vm := range vmr {
		if vm.Template == 1 {
			continue
		}

		if vm.Type != "qemu" {
			continue
		}

		if len(filter) == 0 {
			return vm, nil
		}

		for _, f := range filter {
			ok, err := f(vm)
			if err != nil {
				return nil, err
			}

			if ok {
				return vm, nil
			}
		}
	}

	return nil, ErrVirtualMachineNotFound
}

// GetVMsByFilter returns a VM cluster resource by applying the provided filter functions.
// nolint: dupl
func (c *APIClient) GetVMsByFilter(ctx context.Context, filter ...func(*proxmox.ClusterResource) (bool, error)) (proxmox.ClusterResources, error) {
	vmr, err := c.getResources(ctx, "vm")
	if err != nil {
		return nil, err
	}

	vms := proxmox.ClusterResources{}

	for _, vm := range vmr {
		if vm.Template == 1 {
			continue
		}

		if vm.Type != "qemu" {
			continue
		}

		if len(filter) == 0 {
			vms = append(vms, vm)
		}

		for _, f := range filter {
			ok, err := f(vm)
			if err != nil {
				return nil, err
			}

			if ok {
				vms = append(vms, vm)
			}
		}
	}

	if len(vms) > 0 {
		return vms, nil
	}

	return nil, ErrVirtualMachineNotFound
}

// GetVMTemplateByID returns a VM cluster resource by its ID.
func (c *APIClient) GetVMTemplateByID(ctx context.Context, vmID uint64) (*proxmox.ClusterResource, error) {
	vms, err := c.GetVMTemplatesByFilter(ctx, func(r *proxmox.ClusterResource) (bool, error) {
		return r.VMID == vmID, nil
	})
	if err != nil {
		return nil, err
	}

	if len(vms) == 1 && vms[0].VMID != 0 {
		return vms[0], nil
	}

	return nil, ErrVirtualMachineNotFound
}

// GetVMTemplatesByFilter returns a VM cluster resource by applying the provided filter functions.
// nolint: dupl
func (c *APIClient) GetVMTemplatesByFilter(ctx context.Context, filter ...func(*proxmox.ClusterResource) (bool, error)) (proxmox.ClusterResources, error) {
	vmr, err := c.getResources(ctx, "vm")
	if err != nil {
		return nil, err
	}

	vms := proxmox.ClusterResources{}

	for _, vm := range vmr {
		if vm.Template == 0 {
			continue
		}

		if vm.Type != "qemu" {
			continue
		}

		if len(filter) == 0 {
			vms = append(vms, vm)
		}

		for _, f := range filter {
			ok, err := f(vm)
			if err != nil {
				return nil, err
			}

			if ok {
				vms = append(vms, vm)
			}
		}
	}

	if len(vms) > 0 {
		return vms, nil
	}

	return nil, ErrVirtualMachineTemplateNotFound
}

// GetVMConfig retrieves the configuration of a VM by its ID.
func (c *APIClient) GetVMConfig(ctx context.Context, vmID int) (*proxmox.VirtualMachine, error) {
	vmr, err := c.GetVMByID(ctx, uint64(vmID))
	if err != nil {
		return nil, err
	}

	if vmr.Status == "unknown" { // nolint: goconst
		return nil, ErrVirtualMachineUnreachable
	}

	vm := &proxmox.VirtualMachine{}
	vm.New(c.Client, vmr.Node, vmID)

	if err := vm.Ping(ctx); err != nil {
		return nil, err
	}

	if err := c.Client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/config", vmr.Node, vmID), &vm.VirtualMachineConfig); err != nil {
		return nil, err
	}

	return vm, nil
}

// GetVMTemplateConfig retrieves the configuration of a VM template by its ID.
func (c *APIClient) GetVMTemplateConfig(ctx context.Context, vmID int) (*proxmox.VirtualMachine, error) {
	vmr, err := c.GetVMTemplateByID(ctx, uint64(vmID))
	if err != nil {
		return nil, err
	}

	if vmr.Status == "unknown" {
		return nil, ErrVirtualMachineUnreachable
	}

	vm := &proxmox.VirtualMachine{}
	vm.New(c.Client, vmr.Node, vmID)

	if err := c.Client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/config", vmr.Node, vmID), &vm.VirtualMachineConfig); err != nil {
		return nil, err
	}

	return vm, nil
}

// GetNextID retrieves the next available VM ID.
func (c *APIClient) GetNextID(ctx context.Context, vmid int) (int, error) {
	var ret string

	if _, found := c.lastVMID.Get(strconv.Itoa(vmid)); found {
		return c.GetNextID(ctx, vmid+1)
	}

	data := make(map[string]interface{})
	data["vmid"] = vmid

	if err := c.Client.GetWithParams(ctx, "/cluster/nextid", data, &ret); err != nil {
		if strings.HasPrefix(err.Error(), "bad request: 400 ") {
			return c.GetNextID(ctx, vmid+1)
		}

		return 0, err
	}

	c.lastVMID.SetDefault(strconv.Itoa(vmid), struct{}{})

	return strconv.Atoi(ret)
}
