/*
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
#pragma once

#include "helpers.h"

#define TC_ORIGIN_FLAG 0b00001000

struct bpf_elf_map __section("maps") osm_cki_fib = {
        .type = BPF_MAP_TYPE_LRU_HASH,
        .size_key = sizeof(__u64),
        .size_value = sizeof(struct origin_info),
        .max_elem = 65535,
};

// local_pods stores Pods' ips in current node.
// which can be set by controller.
// only contains injected pods.
struct bpf_elf_map __section("maps") osm_pod_fib = {
        .type = BPF_MAP_TYPE_HASH,
        .size_key = sizeof(__u32) * 4,
        .size_value = sizeof(struct pod_config),
        .max_elem = 1024,
        .pinning = PIN_GLOBAL_NS,
};

// osm_proc_fib stores sidecar's ip address.
struct bpf_elf_map __section("maps") osm_proc_fib = {
        .type = BPF_MAP_TYPE_LRU_HASH,
        .size_key = sizeof(__u32),
        .size_value = sizeof(__u32),
        .max_elem = 1024,
};

// cgroup_ips caches the ip address of each cgroup, which is used to speed up
// the connect process.
struct bpf_elf_map __section("maps") osm_cgr_fib = {
        .type = BPF_MAP_TYPE_LRU_HASH,
        .size_key = sizeof(__u64),
        .size_value = sizeof(struct cgroup_info),
        .max_elem = 1024,
};

struct bpf_elf_map __section("maps") osm_nat_fib = {
        .type = BPF_MAP_TYPE_LRU_HASH,
        .size_key = sizeof(struct pair),
        .size_value = sizeof(struct origin_info),
        .max_elem = 65535,
        .pinning = PIN_GLOBAL_NS,
};

struct bpf_elf_map __section("maps") osm_sock_fib = {
        .type = BPF_MAP_TYPE_SOCKHASH,
        .size_key = sizeof(struct pair),
        .size_value = sizeof(__u32),
        .max_elem = 65535,
};

struct bpf_elf_map __section("maps") osm_mark_fib = {
        .type = BPF_MAP_TYPE_HASH,
        .size_key = sizeof(__u32),
        .size_value = sizeof(__u32) * 4,
        .max_elem = 65535,
};
