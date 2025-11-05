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

	"github.com/luthermonson/go-proxmox"
)

// GetClusterStorage returns the cluster storage resource by name.
func (c *APIClient) GetClusterStorage(ctx context.Context, storage string) (*proxmox.ClusterResource, error) {
	cluster, err := c.Client.Cluster(ctx)
	if err != nil {
		return nil, err
	}

	storageResources, err := cluster.Resources(ctx, "storage")
	if err != nil {
		return nil, err
	}

	for _, resource := range storageResources {
		if resource.Storage == storage {
			return resource, nil
		}
	}

	return nil, ErrNotFound
}

// GetNodeForStorage returns the node name where the storage is available.
func (c *APIClient) GetNodeForStorage(ctx context.Context, storage string) (string, error) {
	cluster, err := c.Client.Cluster(ctx)
	if err != nil {
		return "", err
	}

	storageResources, err := cluster.Resources(ctx, "storage")
	if err != nil {
		return "", err
	}

	for _, resource := range storageResources {
		if resource.Storage == storage && resource.Status == "available" {
			return resource.Node, nil
		}
	}

	return "", ErrNotFound
}

// GetStorageStatus returns the storage status for a given storage on a given node.
func (c *APIClient) GetStorageStatus(ctx context.Context, node string, storage string) (*proxmox.Storage, error) {
	n, err := c.Client.Node(ctx, node)
	if err != nil {
		return nil, err
	}

	st, err := n.Storage(ctx, storage)
	if err != nil {
		return nil, err
	}

	return st, nil
}
