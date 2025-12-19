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

// Package goproxmox implements a proxmox api client.
package goproxmox

import (
	"context"
	"fmt"

	"github.com/luthermonson/go-proxmox"
)

// GetClusterStorage returns the cluster storage resource by name.
func (c *APIClient) GetClusterStorage(ctx context.Context, storage string) (*proxmox.ClusterResource, error) {
	resources, err := c.getResources(ctx, "storage")
	if err != nil {
		return nil, err
	}

	for _, resource := range resources {
		if resource.Storage == storage {
			return resource, nil
		}
	}

	return nil, ErrNotFound
}

// GetNodeForStorage returns the node name where the storage is available.
func (c *APIClient) GetNodeForStorage(ctx context.Context, storage string) (string, error) {
	resources, err := c.getResources(ctx, "storage")
	if err != nil {
		return "", err
	}

	for _, resource := range resources {
		if resource.Storage == storage && resource.Status == "available" {
			return resource.Node, nil
		}
	}

	return "", ErrNotFound
}

// GetStorageStatus returns the storage status for a given storage on a given node.
func (c *APIClient) GetStorageStatus(ctx context.Context, node string, storage string) (st proxmox.Storage, err error) {
	return st, c.Client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/status", node, storage), &st)
}

func (c *APIClient) getResources(ctx context.Context, name string) (proxmox.ClusterResources, error) {
	resources := proxmox.ClusterResources{}

	if v, ok := c.resources.Get(name); ok {
		resources, _ = v.(proxmox.ClusterResources)
	}

	if len(resources) == 0 {
		cluster, err := c.Client.Cluster(ctx)
		if err != nil {
			return nil, err
		}

		resources, err = cluster.Resources(ctx, name)
		if err != nil {
			return nil, err
		}

		c.resources.SetDefault(name, resources)
	}

	return resources, nil
}
