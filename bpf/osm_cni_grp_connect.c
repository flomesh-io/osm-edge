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
#include "headers/cgroup.h"
#include "headers/helpers.h"
#include "headers/maps.h"
#include "headers/mesh.h"
#include <linux/bpf.h>
#include <linux/in.h>

static __u32 outip = 1;

static inline int tcp_connect4(struct bpf_sock_addr *ctx) {
    struct cgroup_info cg_info;
    if (!get_current_cgroup_info(ctx, &cg_info)) {
        return 1;
    }
    if (!cg_info.is_in_mesh) {
        // bypass normal traffic. we only deal pod's
        // traffic managed by mesh.
        return 1;
    }
    __u32 curr_pod_ip;
    __u32 _curr_pod_ip[4];
    set_ipv6(_curr_pod_ip, cg_info.cgroup_ip);
    curr_pod_ip = get_ipv4(_curr_pod_ip);

    if (curr_pod_ip == 0) {
        debugf("get current pod ip error");
    }
    __u64 uid = bpf_get_current_uid_gid() & 0xffffffff;
    __u32 dst_ip = ctx->user_ip4;
    debugf("tcp_connect4 uid: %d pod ip: %pI4 dst ip: %pI4", uid, &curr_pod_ip, &dst_ip);
    if (uid != SIDECAR_USER_ID) {
        if ((dst_ip & 0xff) == 0x7f) {
            debugf("tcp_connect4 [App->Local]: bypass");
            // app call local, bypass.
            return 1;
        }
        __u64 cookie = bpf_get_socket_cookie_addr(ctx);
        // app call other app
        debugf("tcp_connect4 [App->App]: dst ip: %pI4 dst port: %d", &dst_ip,
               bpf_htons(ctx->user_port));

        // we need redirect it to sidecar.
        struct origin_info origin;
        memset(&origin, 0, sizeof(origin));
        set_ipv4(origin.ip, dst_ip);
        origin.port = ctx->user_port;
        origin.flags = 1;
        if (bpf_map_update_elem(&osm_cki_fib, &cookie, &origin, BPF_ANY)) {
            debugf("write osm_cki_fib failed");
            return 0;
        }
        if (curr_pod_ip) {
            struct pod_config *pod = bpf_map_lookup_elem(&osm_pod_fib, _curr_pod_ip);
            if (pod) {
                int exclude = 0;
                IS_EXCLUDE_PORT(pod->exclude_out_ports, ctx->user_port,
                                &exclude);
                if (exclude) {
                    debugf("ignored dest port by exclude_out_ports, ip: "
                           "%pI4, port: %d",
                           &curr_pod_ip, bpf_htons(ctx->user_port));
                    return 1;
                }
                IS_EXCLUDE_IPRANGES(pod->exclude_out_ranges, dst_ip, &exclude);
                debugf("exclude ipranges: %x, exclude: %d",
                       pod->exclude_out_ranges[0].net, exclude);
                if (exclude) {
                    debugf(
                            "ignored dest ranges by exclude_out_ranges, ip: %pI4",
                            &dst_ip);
                    return 1;
                }
                int include = 0;
                IS_INCLUDE_PORT(pod->include_out_ports, ctx->user_port,
                                &include);
                if (!include) {
                    debugf("dest port %d not in pod(%pI4)'s include_out_ports, "
                           "ignored.",
                           bpf_htons(ctx->user_port), &curr_pod_ip);
                    return 1;
                }

                IS_INCLUDE_IPRANGES(pod->include_out_ranges, dst_ip, &include);
                if (!include) {
                    debugf("dest %pI4 not in pod(%pI4)'s include_out_ranges, "
                           "ignored.",
                           &dst_ip, &curr_pod_ip);
                    return 1;
                }
            } else {
                debugf("current pod ip found(%pI4), but can not find pod_info "
                       "from osm_pod_fib",
                       &curr_pod_ip);
            }
            // todo port or ipranges ignore.
            // if we can get the pod ip, we use bind func to bind the pod's ip
            // as the source ip to avoid quaternions conflict of different pods.
            struct sockaddr_in addr = {
                    .sin_addr =
                            {
                                    .s_addr = curr_pod_ip,
                            },
                    .sin_port = 0,
                    .sin_family = 2,
            };
            // todo(kebe7jun) use the following way will cause an error like:
            /*
                578: (07) r2 += -40
                ; if (bpf_bind(ctx, &addr, sizeof(struct sockaddr_in))) {
                579: (bf) r1 = r6
                580: (b7) r3 = 16
                581: (85) call bpf_bind#64
                invalid indirect read from stack R2 off -40+8 size 16
                processed 1136 insns (limit 1000000) max_states_per_insn 1
               total_states 81 peak_states 81 mark_read 20

                libbpf: -- END LOG --
                libbpf: failed to load program 'cgroup/connect4'
                libbpf: failed to load object 'osm_cni_grp_connect.o'
            */
            // addr.sin_addr.s_addr = curr_pod_ip;
            // addr.sin_port = 0;
            // addr.sin_family = 2;
            if (bpf_bind(ctx, &addr, sizeof(struct sockaddr_in))) {
                debugf("bind %pI4 error", &curr_pod_ip);
            }
            ctx->user_ip4 = localhost;
        } else {
            debugf("curr_pod_ip false");
            // if we can not get the pod ip, we rewrite the dest address.
            // The reason we try the IP of the 127.128.0.0/20 segment instead of
            // using 127.0.0.1 directly is to avoid conflicts between the
            // quaternions of different Pods when the quaternions are
            // subsequently processed.
            ctx->user_ip4 = bpf_htonl(0x7f800000 | (outip++));
            if (outip >> 20) {
                outip = 1;
            }
        }
        ctx->user_port = bpf_htons(OUT_REDIRECT_PORT);
        __u32 rewrite_dst_ip = ctx->user_ip4;
        debugf("tcp_connect4 [App->Sidecar]: rewrite dst ip: %pI4, redirect dst port: %d",
               &rewrite_dst_ip, bpf_htons(ctx->user_port));
    } else {
        // from sidecar to others
        __u32 _dst_ip[4];
        set_ipv4(_dst_ip, dst_ip);
        struct pod_config *pod = bpf_map_lookup_elem(&osm_pod_fib, _dst_ip);
        if (!pod) {
            debugf("tcp_connect4 [Sidecar->Sidecar]: uid: %d", uid);
            debugf("tcp_connect4 [Sidecar->Sidecar]: cur pod ip: %pI4 src port: %d", &curr_pod_ip);
            debugf("tcp_connect4 [Sidecar->Sidecar]: dst pod ip: %pI4 dst port: %d", &dst_ip, bpf_htons(ctx->user_port));
            // dst ip is not in this node, bypass
            debugf("tcp_connect4 dest ip: %pI4 not in this node, bypass", &dst_ip);
            return 1;
        }

        // dst ip is in this node, but not the current pod,
        // it is sidecar to sidecar connecting.
        struct origin_info origin;
        memset(&origin, 0, sizeof(origin));
        set_ipv4(origin.ip, dst_ip);
        origin.port = ctx->user_port;

        debugf("tcp_connect4 [Sidecar->Sidecar]: uid: %d", uid);
        debugf("tcp_connect4 [Sidecar->Sidecar]: cur pod ip: %pI4", &curr_pod_ip);
        debugf("tcp_connect4 [Sidecar->Sidecar]: dst pod ip: %pI4 dst port: %d", &dst_ip, bpf_htons(ctx->user_port));

        if (curr_pod_ip) {
            if (curr_pod_ip != dst_ip) {
                // call other pod, need redirect port.
                int exclude = 0;
                IS_EXCLUDE_PORT(pod->exclude_in_ports, ctx->user_port,
                                &exclude);
                if (exclude) {
                    debugf("ignored dest port by exclude_in_ports, ip: %pI4, "
                           "port: %d",
                           &dst_ip, bpf_htons(ctx->user_port));
                    return 1;
                }
                int include = 0;
                IS_INCLUDE_PORT(pod->include_in_ports, ctx->user_port,
                                &include);
                if (!include) {
                    debugf("ignored dest port by include_in_ports, ip: %pI4, "
                           "port: %d",
                           &dst_ip, bpf_htons(ctx->user_port));
                    return 1;
                }
                ctx->user_port = bpf_htons(IN_REDIRECT_PORT);
            }
            origin.flags |= 1;
        } else {
            // can not get current pod ip, we use the legacy mode.

            // u64 bpf_get_current_pid_tgid(void)
            // Return A 64-bit integer containing the current tgid and
            //                 pid, and created as such: current_task->tgid <<
            //                 32
            //                | current_task->pid.
            // pid may be thread id, we should use tgid
            __u32 pid = bpf_get_current_pid_tgid() >> 32; // tgid
            void *curr_ip = bpf_map_lookup_elem(&osm_proc_fib, &pid);
            debugf("tcp_connect4 [Sidecar->Others]: pid: %d, dst ip: %pI4 dst port:%d", pid, &dst_ip,
                   bpf_htons(ctx->user_port));
            if (curr_ip) {
                // sidecar to other sidecar
                if (*(__u32 * )curr_ip != dst_ip) {
                    debugf("tcp_connect4 sidecar to other, rewrite dst port from %d to %d",
                           bpf_htons(ctx->user_port), IN_REDIRECT_PORT);
                    ctx->user_port = bpf_htons(IN_REDIRECT_PORT);
                }
                //origin.flags |= 1;
                origin.flags = 0;
                origin.pid = pid;
                // sidecar to app, no rewrite
                debugf("tcp_connect4 [Sidecar->Others{App}]: sidecar to app, rewrite");
            } else {
                origin.flags = 0;
                origin.pid = pid;
                // sidecar to sidecar
                // try redirect to 15006
                // but it may cause error if it is envoy call self pod,
                // in this case, we can read src and dst ip in sockops,
                // if src is equals dst, it means envoy call self pod,
                // we should reject this traffic in sockops,
                // envoy will create a new connection to self pod.
                debugf("tcp_connect4 [Sidecar->Others{Sidecar}]: sidecar to sidecar, rewrite dst port from %d to %d",
                       bpf_htons(ctx->user_port), IN_REDIRECT_PORT);
                ctx->user_port = bpf_htons(IN_REDIRECT_PORT);
                //debugf("tcp_connect4 [Sidecar->Others{Sidecar}]: sidecar to sidecar, dst port: %d", bpf_htons(ctx->user_port));
            }
        }
        __u64 cookie = bpf_get_socket_cookie_addr(ctx);
        debugf("tcp_connect4 [Sidecar->Others]: call from sidecar container: cookie: %d, dst ip: %pI4, dst port: %d",
               cookie, &dst_ip, bpf_htons(ctx->user_port));
        if (bpf_map_update_elem(&osm_cki_fib, &cookie, &origin, BPF_NOEXIST)) {
            printk("update cookie origin failed");
            return 0;
        }
    }

    return 1;
}

__section("cgroup/connect4") int osm_cni_group_connect4(struct bpf_sock_addr *ctx) {
    switch (ctx->protocol) {
        case IPPROTO_TCP:
            return tcp_connect4(ctx);
        default:
            return 1;
    }
}

char ____license[] __section("license") = "GPL";
int _version __section("version") = 1;
