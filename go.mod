module github.com/sergelogvinov/go-proxmox

go 1.26.1

// replace github.com/luthermonson/go-proxmox => ../go-proxmox-luthermonson
// replace github.com/luthermonson/go-proxmox => github.com/sergelogvinov/go-proxmox-luthermonson v0.0.0-20260101044051-841bcf512414

require (
	github.com/avast/retry-go/v4 v4.7.0
	github.com/luthermonson/go-proxmox v0.4.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.12.1
	go.yaml.in/yaml/v3 v3.0.5
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3
)

require (
	github.com/buger/goterm v1.0.4 // indirect
	github.com/diskfs/go-diskfs v1.9.4 // indirect
	github.com/djherbis/times v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/magefile/mage v1.17.2 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)
