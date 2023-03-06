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
#include "headers/helpers.h"
#include "headers/maps.h"
#include <linux/bpf.h>
#include <linux/in.h>

#define MAX_OPS_BUFF_LENGTH 4096
#define SO_ORIGINAL_DST 80

__section("cgroup/getsockopt") int osm_cni_sock_opt(struct bpf_sockopt *ctx) {
    // currently, eBPF can not deal with optlen more than 4096 bytes, so, we
    // should limit this.
    if (ctx->optlen > MAX_OPS_BUFF_LENGTH) {
        //debugf("optname: %d, force set optlen to %d, original optlen %d is too high",
        //       ctx->optname, MAX_OPS_BUFF_LENGTH, ctx->optlen);
        ctx->optlen = MAX_OPS_BUFF_LENGTH;
    }
    // envoy will call getsockopt with SO_ORIGINAL_DST, we should rewrite it to
    // return original dst info.
    if (ctx->optname != SO_ORIGINAL_DST) {
        return 1;
    }
    debugf("osm_cni_sock_opt ctx->optname:SO_ORIGINAL_DST");
    struct pair p;
    memset(&p, 0, sizeof(p));
    p.dport = bpf_htons(ctx->sk->src_port);
    p.sport = ctx->sk->dst_port;
    struct origin_info *origin;
    switch (ctx->sk->family) {
        case 2: // ipv4
            set_ipv4(p.dip, ctx->sk->src_ip4);
            set_ipv4(p.sip, ctx->sk->dst_ip4);
            __u32 dst_ip4 = get_ipv4(p.dip);
            __u32 src_ip4 = get_ipv4(p.sip);
            debugf("osm_cni_sock_opt src ip4: %pI4 src port: %d", &src_ip4, p.sport);
            debugf("osm_cni_sock_opt dst ip4: %pI4 dst port: %d", &dst_ip4, bpf_htons(p.dport));
            origin = bpf_map_lookup_elem(&osm_nat_fib, &p);
            if (origin) {
                // rewrite original_dst
                ctx->optlen = (__s32)
                sizeof(struct sockaddr_in);
                if ((void *) ((struct sockaddr_in *) ctx->optval + 1) >
                    ctx->optval_end) {
                    printk("optname: %d: invalid getsockopt optval", ctx->optname);
                    return 1;
                }
                ctx->retval = 0;
                struct sockaddr_in sa = {
                        .sin_family = ctx->sk->family,
                        .sin_addr.s_addr = get_ipv4(origin->ip),
                        .sin_port = origin->port,
                };
                *(struct sockaddr_in *) ctx->optval = sa;

                __u32 origin_ip4 = get_ipv4(origin->ip);
                debugf("osm_cni_sock_opt origin dst ip4:%pI4 origin dst port: %d", &origin_ip4, bpf_htons(origin->port));
            } else {
                debugf("osm_cni_sock_opt osm_nat_fib:NOT FOUND");
                //ctx->optlen = (__s32)sizeof(struct sockaddr_in);
                //if ((void *)((struct sockaddr_in *)ctx->optval + 1) >
                //    ctx->optval_end) {
                //    printk("optname: %d: invalid getsockopt optval", ctx->optname);
                //    return 1;
                //}
                //ctx->retval = 0;
                //struct sockaddr_in sa = {
                //    .sin_family = ctx->sk->family,
                //    .sin_addr.s_addr = localhost,
                //    .sin_port = bpf_htons(5000),
                //};
                //*(struct sockaddr_in *)ctx->optval = sa;
                //debugf("osm_cni_sock_opt rewite_origin_ip4:127.0.0.1  rewite_origin_port: %d", 5000);
            }
            break;
    }
    return 1;
}

char ____license[] __section("license") = "GPL";
int _version __section("version") = 1;
