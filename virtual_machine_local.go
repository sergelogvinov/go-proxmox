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
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/luthermonson/go-proxmox"
	yaml "go.yaml.in/yaml/v3"
)

func GetLocalVMConfigByFilter(filter ...func(*proxmox.VirtualMachineConfig) (bool, error)) (int, *proxmox.VirtualMachineConfig, error) {
	entries, err := os.ReadDir("/etc/pve/qemu-server/")
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read qemu-server directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if vmIDStr, ok := strings.CutSuffix(entry.Name(), ".conf"); ok {
			vmID, err := strconv.Atoi(vmIDStr)
			if err != nil {
				continue // Skip non-numeric filenames
			}

			vm, err := GetLocalVMConfig(vmID)
			if err != nil {
				continue // Skip VMs that can't be read
			}

			if len(filter) == 0 {
				return vmID, vm, nil
			}

			for _, filterFunc := range filter {
				match, err := filterFunc(vm)
				if err != nil {
					return 0, nil, fmt.Errorf("filter function error for VM %d: %w", vmID, err)
				}

				if match {
					return vmID, vm, nil
				}
			}
		}
	}

	return 0, nil, ErrVirtualMachineNotFound
}

// GetLocalVMConfig retrieves the configuration of a VM by its ID.
func GetLocalVMConfig(vmID int) (*proxmox.VirtualMachineConfig, error) {
	configPath := fmt.Sprintf("/etc/pve/qemu-server/%d.conf", vmID)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("VM config file not found: %s", configPath)
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read VM config file %s: %w", configPath, err)
	}

	if idx := strings.Index(string(configData), "[PENDING]"); idx != -1 {
		configData = configData[:idx]
	}

	vm := &proxmox.VirtualMachineConfig{}
	if err := yaml.Unmarshal(configData, vm); err != nil {
		return nil, fmt.Errorf("failed to parse VM config for VM %d: %w", vmID, err)
	}

	return vm, nil
}

// GetLocalNextID retrieves the next available VM ID.
func GetLocalNextID(ctx context.Context) (int, error) {
	cmd := exec.CommandContext(ctx, "pvesh", "get", "/cluster/nextid")

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	vmID, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, err
	}

	return vmID, nil
}

func CreateLocalVM(ctx context.Context, vmID int, options map[string]any) error {
	args := make([]string, 0, 2+len(options))
	args = append(args, "create", strconv.Itoa(vmID))

	for key, value := range options {
		args = append(args, fmt.Sprintf("--%s", key), fmt.Sprintf("%v", value))
	}

	cmd := exec.CommandContext(ctx, "qm", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create VM %d: %w, output: %s", vmID, err, string(output))
	}

	return nil
}

func DeleteLocalVM(ctx context.Context, vmID int) error {
	cmd := exec.CommandContext(ctx, "qm", "destroy", strconv.Itoa(vmID))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete VM %d: %w, output: %s", vmID, err, string(output))
	}

	return nil
}

func UpdateLocalVM(ctx context.Context, vmID int, options map[string]any) error {
	configPath := fmt.Sprintf("/etc/pve/qemu-server/%d.conf", vmID)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("VM config file not found: %s", configPath)
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read VM config file %s: %w", configPath, err)
	}

	if idx := strings.Index(string(configData), "[PENDING]"); idx != -1 {
		configData = configData[:idx]
	}

	vmOptions := map[string]any{}
	if err := yaml.Unmarshal(configData, vmOptions); err != nil {
		return fmt.Errorf("failed to parse VM config for VM %d: %w", vmID, err)
	}

	updateOptions := map[string]any{}

	for key, value := range options {
		if v, ok := vmOptions[key]; ok && v == value {
			continue
		}

		updateOptions[key] = value
	}

	if len(updateOptions) == 0 {
		return nil
	}

	args := make([]string, 0, 2+len(updateOptions))
	args = append(args, "set", strconv.Itoa(vmID))

	for key, value := range updateOptions {
		args = append(args, fmt.Sprintf("--%s", key), fmt.Sprintf("%v", value))
	}

	cmd := exec.CommandContext(ctx, "qm", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update VM %d: %w, output: %s", vmID, err, string(output))
	}

	return nil
}
