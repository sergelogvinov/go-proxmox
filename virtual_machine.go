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

// FindVMByID tries to find a VM by its ID on the whole cluster.
func (c *APIClient) FindVMByID(ctx context.Context, vmID uint64) (*proxmox.ClusterResource, error) {
	resources, err := c.getResources(ctx, "vm")
	if err != nil {
		return nil, err
	}

	for _, vm := range resources {
		if vm.Template == 1 {
			continue
		}

		if vm.VMID == vmID {
			return vm, nil
		}
	}

	return nil, ErrVirtualMachineNotFound
}

// FindVMByName tries to find a VMID by its name
func (c *APIClient) FindVMByName(ctx context.Context, name string) (vmID int, err error) {
	resources, err := c.getResources(ctx, "vm")
	if err != nil {
		return 0, err
	}

	for _, vm := range resources {
		if vm.Template == 1 {
			continue
		}

		if vm.Name == name {
			return int(vm.VMID), nil
		}
	}

	return 0, ErrVirtualMachineNotFound
}

// FindVMByFilter tries to find a VMID by applying filter functions
func (c *APIClient) FindVMByFilter(ctx context.Context, filter ...func(*proxmox.ClusterResource) (bool, error)) (vmID int, err error) {
	resources, err := c.getResources(ctx, "vm")
	if err != nil {
		return 0, err
	}

	for _, vm := range resources {
		if vm.Template == 1 {
			continue
		}

		for _, f := range filter {
			ok, err := f(vm)
			if err != nil {
				return 0, err
			}

			if ok {
				return int(vm.VMID), nil
			}
		}
	}

	return 0, ErrVirtualMachineNotFound
}

// FindVMTemplateByName tries to find a VMID by its name
func (c *APIClient) FindVMTemplateByName(ctx context.Context, zone, name string) (vmID int, err error) {
	resources, err := c.getResources(ctx, "vm")
	if err != nil {
		return 0, err
	}

	for _, vm := range resources {
		if vm.Template == 0 {
			continue
		}

		if vm.Node == zone && vm.Name == name {
			return int(vm.VMID), nil
		}
	}

	if vmID == 0 {
		return 0, ErrVirtualMachineTemplateNotFound
	}

	return vmID, nil
}

// GetVMStatus retrieves the status of a VM by its ID.
func (c *APIClient) GetVMStatus(ctx context.Context, vmid int) (*proxmox.ClusterResource, error) {
	resources, err := c.getResources(ctx, "vm")
	if err != nil {
		return nil, err
	}

	for _, vm := range resources {
		if vm.Template == 1 {
			continue
		}

		if vm.VMID == uint64(vmid) {
			return vm, nil
		}
	}

	return nil, ErrVirtualMachineNotFound
}

// GetVMConfig retrieves the configuration of a VM by its ID.
func (c *APIClient) GetVMConfig(ctx context.Context, vmID int) (*proxmox.VirtualMachine, error) {
	vmr, err := c.GetVMStatus(ctx, vmID)
	if err != nil {
		return nil, err
	}

	// if vmr.Status == "unknown" {
	// 	return nil, ErrVirtualMachineUnreachable
	// }

	node, err := c.Node(ctx, vmr.Node)
	if err != nil {
		return nil, err
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return nil, err
	}

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

// StartVMByID starts a VM by its ID.
func (c *APIClient) StartVMByID(ctx context.Context, nodeName string, vmID int) (*proxmox.VirtualMachine, error) {
	node, err := c.Node(ctx, nodeName)
	if err != nil {
		return nil, fmt.Errorf("unable to find node with name %s: %w", nodeName, err)
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return nil, fmt.Errorf("unable to find vm with id %d: %w", vmID, err)
	}

	if _, err := vm.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start vm %d: %v", vmID, err)
	}

	return vm, nil
}

// DeleteVMByID deletes a VM by its ID.
func (c *APIClient) DeleteVMByID(ctx context.Context, nodeName string, vmID int) error {
	node, err := c.Node(ctx, nodeName)
	if err != nil {
		return fmt.Errorf("unable to find node with name %s: %w", nodeName, err)
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return fmt.Errorf("unable to find vm with id %d: %w", vmID, err)
	}

	if vm.IsRunning() {
		if _, err := vm.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop vm %d: %v", vmID, err)
		}
	}

	if _, err := vm.Delete(ctx); err != nil {
		return fmt.Errorf("cannot delete vm with id %d: %w", vmID, err)
	}

	c.flushResources("vm")
	c.lastVMID.SetDefault(strconv.Itoa(vmID), struct{}{})

	return nil
}

// MigrateVMByID migrates a VM to another node by its ID.
func (c *APIClient) MigrateVMByID(ctx context.Context, vmID int, dstNode string, online bool) error {
	vm, err := c.FindVMByID(ctx, uint64(vmID))
	if err != nil {
		return err
	}

	params := &proxmox.VirtualMachineMigrateOptions{
		Target: dstNode,
		Online: proxmox.IntOrBool(online),
	}

	var upid proxmox.UPID
	if err = c.Client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/migrate", vm.Node, vm.VMID), params, &upid); err != nil {
		return err
	}

	task := proxmox.NewTask(upid, c.Client)
	if task != nil {
		if err = task.WaitFor(ctx, 5*60); err != nil {
			return fmt.Errorf("unable to migrate virtual machine: %w", err)
		}

		if task.IsFailed {
			return fmt.Errorf("unable to migrate virtual machine: %s", task.ExitStatus)
		}
	}

	return nil
}

// CreateVM creates a new VM on the specified node with the given configuration.
func (c *APIClient) CreateVM(ctx context.Context, node string, vm map[string]interface{}) error {
	var upid proxmox.UPID

	if err := c.Post(ctx, fmt.Sprintf("/nodes/%s/qemu", node), &vm, &upid); nil != err {
		return fmt.Errorf("unable to create virtual machine: %w", err)
	}

	task := proxmox.NewTask(upid, c.Client)
	if err := task.WaitFor(ctx, 5*60); err != nil {
		return fmt.Errorf("unable to create virtual machine: %w", err)
	}

	if task.IsFailed {
		return fmt.Errorf("unable to create virtual machine: %s", task.ExitStatus)
	}

	return nil
}

// CloneVM clones a VM template to create a new VM with the specified options.
func (c *APIClient) CloneVM(ctx context.Context, templateID int, options VMCloneRequest) (int, error) {
	node, err := c.Node(ctx, options.Node)
	if err != nil {
		return 0, fmt.Errorf("unable to find node with name %s: %w", options.Node, err)
	}

	vmTemplate, err := node.VirtualMachine(ctx, templateID)
	if err != nil {
		return 0, fmt.Errorf("unable to find vm with id %d: %w", templateID, err)
	}

	vmCloneOptions := proxmox.VirtualMachineCloneOptions{
		NewID:       options.NewID,
		Description: options.Description,
		Full:        options.Full,
		Name:        options.Name,
		Pool:        options.Pool,
		Storage:     options.Storage,
	}

	newid, task, err := vmTemplate.Clone(ctx, &vmCloneOptions)
	if err != nil {
		return 0, fmt.Errorf("failed to clone vm template %d: %v", templateID, err)
	}

	if err := task.WaitFor(ctx, 5*60); err != nil {
		return 0, fmt.Errorf("unable to clone virtual machine: %w", err)
	}

	if task.IsFailed {
		return 0, fmt.Errorf("unable to clone virtual machine: %s", task.ExitStatus)
	}

	c.flushResources("vm")

	vm, err := node.VirtualMachine(ctx, newid)
	if err != nil {
		return 0, fmt.Errorf("failed to get vm %d: %v", newid, err)
	}

	if _, err = vm.ResizeDisk(ctx, "scsi0", options.DiskSize); err != nil {
		return 0, fmt.Errorf("failed to resize disk for vm %d: %v", newid, err)
	}

	var vmOptions []proxmox.VirtualMachineOption

	if options.CPU != 0 {
		vmOptions = append(vmOptions, proxmox.VirtualMachineOption{Name: "cores", Value: fmt.Sprintf("%d", options.CPU)})
	}

	if options.Memory != 0 {
		vmOptions = append(vmOptions, proxmox.VirtualMachineOption{Name: "memory", Value: fmt.Sprintf("%d", options.Memory)})
	}

	if options.Tags != "" {
		vmOptions = append(vmOptions, proxmox.VirtualMachineOption{Name: "tags", Value: options.Tags})
	}

	vmOptions = applyInstanceSMBIOS(vm, options, vmOptions)
	vmOptions = applyInstanceOptimization(vm, options, vmOptions)

	if len(vmOptions) > 0 {
		_, err := vm.Config(ctx, vmOptions...)
		if err != nil {
			return 0, fmt.Errorf("unable to configure vm: %w", err)
		}
	}

	return newid, err
}

// RegenerateVMCloudInit regenerates the Cloud-Init configuration for a VM.
func (c *APIClient) RegenerateVMCloudInit(ctx context.Context, node string, vmID int) error {
	if err := c.Put(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/cloudinit", node, vmID), map[string]string{
		"node": node,
		"vmid": fmt.Sprintf("%d", vmID),
	}, nil); err != nil {
		return err
	}

	return nil
}
