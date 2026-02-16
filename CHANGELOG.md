# Changelog

## [0.2.0](https://github.com/sergelogvinov/go-proxmox/compare/v0.1.0...v0.2.0) (2026-02-16)


### Features

* add local VM management ([e7fe4fa](https://github.com/sergelogvinov/go-proxmox/commit/e7fe4fa0b5cc387b0abe35ad8cd37d97ca88ddf5))
* add VM update func ([6c814d1](https://github.com/sergelogvinov/go-proxmox/commit/6c814d1396e00c6090049f356b4105c7504166a5))
* check vm status ([63ad3d6](https://github.com/sergelogvinov/go-proxmox/commit/63ad3d61b87a076f02148d050f4127e0797699fe))
* numa nodes memory allocation ([1bbce55](https://github.com/sergelogvinov/go-proxmox/commit/1bbce55952c6ce32c5fa25a884e8c93851a187f0))
* numa nodes struct ([3f5457e](https://github.com/sergelogvinov/go-proxmox/commit/3f5457e4c5395a84ac262e92a2fa54c299932db6))
* return disk name from CreateVMDisk ([a813c5d](https://github.com/sergelogvinov/go-proxmox/commit/a813c5df5a79954ab3e9407af8378289aec7d22d))
* vm creation verification ([58a517f](https://github.com/sergelogvinov/go-proxmox/commit/58a517f3ee94a961749fa54c6535144972591ddd))


### Bug Fixes

* cloneVM return zero vmid on error ([24982c4](https://github.com/sergelogvinov/go-proxmox/commit/24982c417a2131efed0bc940269bc36a089c33a1))
* flush cache ([462ae25](https://github.com/sergelogvinov/go-proxmox/commit/462ae2542c15f6a0d0341c308d5a7e7df4f1809e))
* flush cache ([c077f53](https://github.com/sergelogvinov/go-proxmox/commit/c077f53b913b1501960db343080425ce7ca37969))
* flush cache ([30957ea](https://github.com/sergelogvinov/go-proxmox/commit/30957ea23dc0beaf96e88686e243d1f6bb8863a2))
* flush cache ([3ff0440](https://github.com/sergelogvinov/go-proxmox/commit/3ff0440b9fb6f0b86619853a75d296598c7e6d76))
* numa index in VM ([198293a](https://github.com/sergelogvinov/go-proxmox/commit/198293aba58a2b22389a714c731a8320b40339fd))
* numa nodes memory allocation ([36e1fe6](https://github.com/sergelogvinov/go-proxmox/commit/36e1fe604b60c0067d2acb608fdae172222386ba))
* numa nodes memory allocation ([ace2202](https://github.com/sergelogvinov/go-proxmox/commit/ace220291bbc3f392136d34838ce16596523ea10))
* skip lxc containers ([f7532aa](https://github.com/sergelogvinov/go-proxmox/commit/f7532aa8ce818dc862f61d2d30b4b4e8553bfea9))

## [0.1.0](https://github.com/sergelogvinov/go-proxmox/compare/v0.0.1...v0.1.0) (2026-01-04)


### Features

* add hostpci ([52881d5](https://github.com/sergelogvinov/go-proxmox/commit/52881d59a6c5043752a726ecfdab55a151b32ec4))
* add qemu guest agent ([b003ecb](https://github.com/sergelogvinov/go-proxmox/commit/b003ecb58e030479b18ba7481d1f5a59b651def7))
* add storage management functions ([ca66549](https://github.com/sergelogvinov/go-proxmox/commit/ca66549d2e5bcd5c8aab6412c0a878e7f4cccf2c))
* add vm smbios func ([f1b40a5](https://github.com/sergelogvinov/go-proxmox/commit/f1b40a5b6d059195b0d1ea8be8075bbe30eb8d15))
* cache for cluster resources ([654365b](https://github.com/sergelogvinov/go-proxmox/commit/654365b267dadb305b7ed6264115fc1c0f5c5b1f))
* cluster ha-groups ([13b553e](https://github.com/sergelogvinov/go-proxmox/commit/13b553e5a6c1ec9a3b7baeb9b101a3f267dc43be))
* find virtual machines funcs ([52c839b](https://github.com/sergelogvinov/go-proxmox/commit/52c839b47048af2f790fd839f1a33c90e6717252))
* pool option ([0b0ddc3](https://github.com/sergelogvinov/go-proxmox/commit/0b0ddc3e8dec8e84026a5f383fa5418f2046151f))
* prefer cluster resource ([741071d](https://github.com/sergelogvinov/go-proxmox/commit/741071df5bbe3cd4679a5bf6f913bde981b9f9a8))
* regenerate cloudinit ([1af35c7](https://github.com/sergelogvinov/go-proxmox/commit/1af35c793934379193cb27f38b9f6b17cd820b38))
* unreachable virtual machine error ([70a3eea](https://github.com/sergelogvinov/go-proxmox/commit/70a3eea3125a2ae68b770ee028687b38092c3026))


### Bug Fixes

* check vm status ([b240815](https://github.com/sergelogvinov/go-proxmox/commit/b2408158419a65f8c6cb1f9cac3d78feaacca68a))
* error message ([dc287d1](https://github.com/sergelogvinov/go-proxmox/commit/dc287d182403ffd1a63e5b2b034461d196127320))
* flush cache ([e5bbe2b](https://github.com/sergelogvinov/go-proxmox/commit/e5bbe2b0806daf184d58029513d0c98f57cab73f))
* set release version ([e0ee5a7](https://github.com/sergelogvinov/go-proxmox/commit/e0ee5a7e85372fd3a88580a2adc4d6ddd17ecabd))
* unmarshal int/string list ([53535c7](https://github.com/sergelogvinov/go-proxmox/commit/53535c777228dde473aa8c51768989ddc4ac2418))
* wait cloning of vm ([2097684](https://github.com/sergelogvinov/go-proxmox/commit/2097684ff6c8a0e224a76ed0f6bd0a9513fe0440))
