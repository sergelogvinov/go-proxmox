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
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/patrickmn/go-cache"
)

// GetHAGroupList retrieves the list of HA groups in the cluster.
func (c *APIClient) GetHAGroupList(ctx context.Context) (groups []*HAGroup, err error) {
	err = c.Get(ctx, "/cluster/ha/groups", &groups)
	if nil != err {
		return nil, err
	}

	return groups, nil
}

func (c *APIClient) getResources(ctx context.Context, name string) (proxmox.ClusterResources, error) {
	resources := proxmox.ClusterResources{}

	if v, ok := c.resources.Get(name); ok {
		resources, _ = v.(proxmox.ClusterResources)
	}

	ttl := cache.DefaultExpiration

	switch name {
	case "vm":
		ttl = time.Second * 5
	case "storage":
		ttl = time.Minute
	}

	if len(resources) == 0 {
		if err := c.Get(ctx, fmt.Sprintf("/cluster/resources?type=%s", name), &resources); err != nil {
			return nil, fmt.Errorf("could not list cluster resources: %w", err)
		}

		c.resources.Set(name, resources, ttl)
	}

	return resources, nil
}
