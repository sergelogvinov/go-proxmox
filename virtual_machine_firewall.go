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
	"reflect"

	"github.com/luthermonson/go-proxmox"
)

func (c *APIClient) CreateVMFirewallRules(ctx context.Context, vmID int, nodeName string, rules []*proxmox.FirewallRule) error {
	node, err := c.Node(ctx, nodeName)
	if err != nil {
		return fmt.Errorf("unable to find node with name %s: %w", nodeName, err)
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return fmt.Errorf("unable to find vm with id %d: %w", vmID, err)
	}

	if len(rules) > 0 {
		vmOptions, err := vm.FirewallOptionGet(ctx)
		if err != nil {
			return fmt.Errorf("failed to get firewall options for vm %d: %v", vmID, err)
		}

		if vmOptions == nil {
			vmOptions = &proxmox.FirewallVirtualMachineOption{
				Enable:    false,
				Dhcp:      true,
				Ipfilter:  false,
				PolicyIn:  "DROP",
				PolicyOut: "ACCEPT",
			}
		}

		vmOptions.Enable = true
		vmOptions.PolicyIn = "DROP"
		if err := vm.FirewallOptionSet(ctx, vmOptions); err != nil {
			return fmt.Errorf("failed to set firewall options for vm %d: %v", vmID, err)
		}

		for _, rule := range rules {
			if err := vm.FirewallRulesCreate(ctx, rule); err != nil {
				return fmt.Errorf("failed to set firewall rule for vm %d: %v", vmID, err)
			}
		}
	}

	return nil
}

func (c *APIClient) UpdateVMFirewallRules(ctx context.Context, vmID int, nodeName string, rules []*proxmox.FirewallRule) error {
	node, err := c.Node(ctx, nodeName)
	if err != nil {
		return fmt.Errorf("unable to find node with name %s: %w", nodeName, err)
	}

	vm, err := node.VirtualMachine(ctx, vmID)
	if err != nil {
		return fmt.Errorf("unable to find vm with id %d: %w", vmID, err)
	}

	oldRules, err := vm.FirewallGetRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to get firewall rules for vm %d: %v", vmID, err)
	}

	n := len(oldRules)
	if n < len(rules) {
		n = len(rules)
	}

	for i := range n {
		switch {
		case i < len(oldRules) && i < len(rules) && !reflect.DeepEqual(oldRules[i], rules[i]):
			if err := vm.FirewallRulesUpdate(ctx, rules[i]); err != nil {
				return fmt.Errorf("failed to update firewall rule for vm %d: %v", vmID, err)
			}
		case i < len(oldRules) && i >= len(rules):
			if err := vm.FirewallRulesDelete(ctx, i); err != nil {
				return fmt.Errorf("failed to delete old firewall rule for vm %d: %v", vmID, err)
			}
		case i >= len(oldRules) && i < len(rules):
			if err := vm.FirewallRulesCreate(ctx, rules[i]); err != nil {
				return fmt.Errorf("failed to create new firewall rule for vm %d: %v", vmID, err)
			}
		}
	}

	return nil
}
