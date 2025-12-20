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
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/patrickmn/go-cache"
)

// APIClient Proxmox API client object.
type APIClient struct {
	*proxmox.Client

	lastVMID  *cache.Cache
	resources *cache.Cache
}

// NewAPIClient initializes a GO-Proxmox API client.
func NewAPIClient(url string, options ...proxmox.Option) (*APIClient, error) {
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	client := proxmox.NewClient(url, options...)

	// _, err := client.Version(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to initialize proxmox api client: %w", err)
	// }

	return &APIClient{
		Client:    client,
		lastVMID:  cache.New(5*time.Minute, 10*time.Minute),
		resources: cache.New(1*time.Minute, 10*time.Minute),
	}, nil
}
