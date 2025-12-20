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

	"github.com/luthermonson/go-proxmox"
)

// CreateVMDisk creates a new disk for the virtual machine.
func (c *APIClient) CreateVMDisk(ctx context.Context, vmid int, node string, storage string, disk string, sizeBytes int64) error {
	params := make(map[string]interface{})
	params["vmid"] = vmid
	params["node"] = node
	params["storage"] = storage
	params["filename"] = disk
	params["size"] = fmt.Sprintf("%d", sizeBytes/1024)

	err := c.Client.Post(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content", node, storage), params, nil)
	if err != nil {
		return fmt.Errorf("unable to create disk for virtual machine: %w", err)
	}

	return nil
}

// DeleteVMDisk deletes a disk from the virtual machine.
func (c *APIClient) DeleteVMDisk(ctx context.Context, node string, storage string, disk string) error {
	var upid proxmox.UPID
	if err := c.Client.Delete(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content/%s", node, storage, disk), &upid); err != nil {
		return err
	}

	task := proxmox.NewTask(upid, c.Client)
	if task != nil {
		if err := task.WaitFor(ctx, 30); err != nil {
			return fmt.Errorf("unable to delete virtual machine disk: %w", err)
		}

		if task.IsFailed {
			return fmt.Errorf("unable to delete virtual machine disk: %s", task.ExitStatus)
		}
	}

	return nil
}

// AttachVMDisk attaches an existing disk to the virtual machine.
func (c *APIClient) AttachVMDisk(ctx context.Context, vmID int, device, disk string) error {
	vmr, err := c.GetVMStatus(ctx, vmID)
	if err != nil {
		return err
	}

	node, err := c.Node(ctx, vmr.Node)
	if err != nil {
		return err
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return err
	}

	vmOptions := proxmox.VirtualMachineOption{
		Name:  device,
		Value: disk,
	}

	task, err := vm.Config(ctx, vmOptions)
	if err != nil {
		return fmt.Errorf("unable to attach disk: %v, options=%+v", err, vmOptions)
	}

	if task != nil {
		if err = task.WaitFor(ctx, 5*60); err != nil {
			return fmt.Errorf("unable to attach virtual machine disk: %w", err)
		}

		if task.IsFailed {
			return fmt.Errorf("unable to attach virtual machine disk: %s", task.ExitStatus)
		}
	}

	return nil
}

// DetachVMDisk detaches a disk from the virtual machine.
func (c *APIClient) DetachVMDisk(ctx context.Context, vmID int, device string) error {
	vmr, err := c.GetVMStatus(ctx, vmID)
	if err != nil {
		return err
	}

	node, err := c.Node(ctx, vmr.Node)
	if err != nil {
		return err
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return err
	}

	task, err := vm.UnlinkDisk(ctx, device, false)
	if err != nil {
		return fmt.Errorf("failed to unlink disk: %v", err)
	}

	if task != nil {
		if err := task.WaitFor(ctx, 5*60); err != nil {
			return fmt.Errorf("unable to detach virtual machine disk: %w", err)
		}

		if task.IsFailed {
			return fmt.Errorf("unable to detach virtual machine disk: %s", task.ExitStatus)
		}
	}

	return nil
}

// ResizeVMDisk resizes a disk for the virtual machine.
func (c *APIClient) ResizeVMDisk(ctx context.Context, vmID int, node, disk, size string) error {
	n, err := c.Client.Node(ctx, node)
	if err != nil {
		return fmt.Errorf("unable to find node with name %s: %w", node, err)
	}

	vm, err := n.VirtualMachine(ctx, vmID)
	if err != nil {
		return err
	}

	task, err := vm.ResizeDisk(ctx, disk, size)
	if err != nil {
		return fmt.Errorf("unable to resize virtual machine disk: %w", err)
	}

	if task == nil {
		return nil
	}

	if err := task.WaitFor(ctx, 5*60); err != nil {
		return fmt.Errorf("unable to resize virtual machine disk: %w", err)
	}

	if task.IsFailed {
		return fmt.Errorf("unable to resize virtual machine disk: %s", task.ExitStatus)
	}

	return nil
}
