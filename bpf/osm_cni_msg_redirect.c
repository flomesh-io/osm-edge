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

__section("sk_msg") int osm_cni_msg_redirect(struct sk_msg_md *msg) {
    struct pair p;
    memset(&p, 0, sizeof(p));
    p.dport = bpf_htons(msg->local_port);
    p.sport = msg->remote_port >> 16;

    switch (msg->family) {
        case 2:
            // ipv4
            set_ipv4(p.dip, msg->local_ip4);
            set_ipv4(p.sip, msg->remote_ip4);
            break;
    }

    __u32 local_ip4 = get_ipv4(p.dip);
    __u32 remote_ip4 = get_ipv4(p.sip);
    debugf("osm_cni_msg_redirect src ip4: %pI4 -> dst ip4: %pI4", &local_ip4, &remote_ip4);
    debugf("osm_cni_msg_redirect src port:%d -> dst port: %d", p.dport, bpf_htons(p.sport));

    long ret = bpf_msg_redirect_hash(msg, &osm_sock_fib, &p, BPF_F_INGRESS);
    if (ret)
        debugf("osm_cni_msg_redirect redirect %d bytes with eBPF successfully", msg->size);
    return 1;
}

char ____license[] __section("license") = "GPL";
int _version __section("version") = 1;
