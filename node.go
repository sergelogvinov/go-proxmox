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

// GetNodeList returns a list of all node names in the cluster.
func (c *APIClient) GetNodeList(ctx context.Context) ([]string, error) {
	ns, err := c.Client.Nodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node list: %v", err)
	}

	nodeList := []string{}

	for _, item := range ns {
		if node := item.Node; node != "" {
			nodeList = append(nodeList, node)
		}
	}

	return nodeList, nil
}

// GetNodeListByFilter get cluster node resources by applying the provided filter functions.
func (c *APIClient) GetNodeListByFilter(ctx context.Context, filter ...func(*proxmox.ClusterResource) (bool, error)) (nodes proxmox.ClusterResources, err error) {
	resources, err := c.getResources(ctx, "node")
	if err != nil {
		return nil, err
	}

	if len(filter) == 0 {
		return resources, nil
	}

	for _, node := range resources {
		for _, f := range filter {
			ok, err := f(node)
			if err != nil {
				return nil, err
			}

			if ok {
				nodes = append(nodes, node)
			}
		}
	}

	return nodes, nil
}
