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

	"github.com/luthermonson/go-proxmox"
)

// StartVMByID starts a VM by its ID.
func (c *APIClient) StartVMByID(ctx context.Context, nodeName string, vmID int) (*proxmox.VirtualMachine, error) {
	vm := &proxmox.VirtualMachine{}
	vm.New(c.Client, nodeName, vmID)

	if err := vm.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to find vm with id %d: %w", vmID, err)
	}

	defer func() {
		c.flushResources("vm")
	}()

	task, err := vm.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start vm %d: %v", vmID, err)
	}

	if task != nil {
		if err = task.WaitFor(ctx, 60); err != nil {
			return nil, fmt.Errorf("unable to start virtual machine: %w", err)
		}

		if task.IsFailed {
			return nil, fmt.Errorf("unable to start virtual machine: %s", task.ExitStatus)
		}
	}

	if err := c.Client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/config", nodeName, vmID), &vm.VirtualMachineConfig); err != nil {
		return nil, err
	}

	return vm, nil
}

// DeleteVMByID deletes a VM by its ID.
func (c *APIClient) DeleteVMByID(ctx context.Context, nodeName string, vmID int) error {
	vm := &proxmox.VirtualMachine{}
	vm.New(c.Client, nodeName, vmID)

	if err := vm.Ping(ctx); err != nil {
		return fmt.Errorf("unable to find vm with id %d: %w", vmID, err)
	}

	if vm.IsRunning() {
		if _, err := vm.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop vm %d: %v", vmID, err)
		}
	}

	defer func() {
		c.flushResources("vm")
	}()

	if _, err := vm.Delete(ctx); err != nil {
		return fmt.Errorf("cannot delete vm with id %d: %w", vmID, err)
	}

	c.lastVMID.SetDefault(strconv.Itoa(vmID), struct{}{})

	return nil
}

// MigrateVMByID migrates a VM to another node by its ID.
func (c *APIClient) MigrateVMByID(ctx context.Context, vmID int, dstNode string, online bool) error {
	vmr, err := c.GetVMByID(ctx, uint64(vmID))
	if err != nil {
		return err
	}

	defer func() {
		c.flushResources("vm")
	}()

	params := &proxmox.VirtualMachineMigrateOptions{
		Target: dstNode,
		Online: proxmox.IntOrBool(online),
	}

	var upid proxmox.UPID
	if err = c.Client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/migrate", vmr.Node, vmr.VMID), params, &upid); err != nil {
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
func (c *APIClient) CreateVM(ctx context.Context, node string, options map[string]interface{}) error {
	var upid proxmox.UPID

	defer func() {
		c.flushResources("vm")
	}()

	template := options["template"] == 1
	if template {
		delete(options, "template")
	}

	if err := c.Post(ctx, fmt.Sprintf("/nodes/%s/qemu", node), &options, &upid); nil != err {
		return fmt.Errorf("unable to create virtual machine: %w", err)
	}

	task := proxmox.NewTask(upid, c.Client)
	if err := task.WaitFor(ctx, 5*60); err != nil {
		return fmt.Errorf("unable to create virtual machine: %w", err)
	}

	if task.IsFailed {
		return fmt.Errorf("unable to create virtual machine: %s", task.ExitStatus)
	}

	c.flushResources("vm")

	if template {
		if err := c.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/template", node, options["vmid"]), nil, &upid); nil != err {
			return fmt.Errorf("unable to create template of virtual machine: %w", err)
		}

		task := proxmox.NewTask(upid, c.Client)
		if err := task.WaitFor(ctx, 60); err != nil {
			return fmt.Errorf("unable to convert to template of virtual machine: %w", err)
		}

		if task.IsFailed {
			return fmt.Errorf("unable to convert to template of virtual machine: %s", task.ExitStatus)
		}
	}

	return nil
}

// UpdateVMByID updates an existing VM on the specified node with the given configuration.
func (c *APIClient) UpdateVMByID(ctx context.Context, nodeName string, vmID int, options map[string]interface{}) error {
	vm := &proxmox.VirtualMachine{}
	vm.New(c.Client, nodeName, vmID)

	if err := c.Client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/config", nodeName, vmID), &vm.VirtualMachineConfig); err != nil {
		return err
	}

	vmOptions := getVMOptionsToApply(vm.VirtualMachineConfig, options)
	if len(vmOptions) == 0 {
		return nil
	}

	defer func() {
		c.flushResources("vm")
	}()

	task, err := vm.Config(ctx, vmOptions...)
	if err != nil {
		return fmt.Errorf("unable to configure vm: %w", err)
	}

	if task != nil {
		if err = task.WaitFor(ctx, 5*60); err != nil {
			return fmt.Errorf("unable to configure virtual machine: %w", err)
		}

		if task.IsFailed {
			return fmt.Errorf("unable to configure virtual machine: %s", task.ExitStatus)
		}
	}

	return nil
}

// CloneVM clones a VM template to create a new VM with the specified options.
func (c *APIClient) CloneVM(ctx context.Context, templateID int, options VMCloneRequest) (int, error) {
	vmTemplate := &proxmox.VirtualMachine{}
	vmTemplate.New(c.Client, options.Node, templateID)

	if err := vmTemplate.Ping(ctx); err != nil {
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

	defer func() {
		c.flushResources("vm")
	}()

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

	vm := &proxmox.VirtualMachine{}
	vm.New(c.Client, options.Node, newid)

	if err := vm.Ping(ctx); err != nil {
		return 0, fmt.Errorf("failed to get vm %d: %v", newid, err)
	}

	// FIXME: remove hardcoded disk name
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
		task, err := vm.Config(ctx, vmOptions...)
		if err != nil {
			return 0, fmt.Errorf("unable to configure vm: %w", err)
		}

		if task != nil {
			if err = task.WaitFor(ctx, 5*60); err != nil {
				return 0, fmt.Errorf("unable to configure virtual machine: %w", err)
			}

			if task.IsFailed {
				return 0, fmt.Errorf("unable to configure virtual machine: %s", task.ExitStatus)
			}
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
