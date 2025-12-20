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

// GetNodesForStorage returns the node name where the storage is available.
func (c *APIClient) GetNodesForStorage(ctx context.Context, storage string) ([]string, error) {
	resources, err := c.getResources(ctx, "storage")
	if err != nil {
		return nil, err
	}

	nodes := []string{}

	for _, resource := range resources {
		if resource.Storage == storage && resource.Status == "available" {
			nodes = append(nodes, resource.Node)
		}
	}

	if len(nodes) == 0 {
		return nil, ErrNotFound
	}

	return nodes, nil
}

// GetStorageStatus returns the storage status for a given storage on a given node.
func (c *APIClient) GetStorageStatus(ctx context.Context, node string, storage string) (st proxmox.Storage, err error) {
	return st, c.Client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/status", node, storage), &st)
}

// GetStorageContent returns the storage content for a given storage on a given node.
func (c *APIClient) GetStorageContent(ctx context.Context, node string, storage string) (content []*proxmox.StorageContent, err error) {
	return content, c.Client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content", node, storage), &content)
}
