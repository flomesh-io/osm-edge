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
#define SOCK_IP_MARK_PORT 15050

#ifndef OUT_REDIRECT_PORT
#define OUT_REDIRECT_PORT 15001
#endif

#ifndef IN_REDIRECT_PORT
#define IN_REDIRECT_PORT 15003
#endif

#ifndef SIDECAR_USER_ID
#define SIDECAR_USER_ID 1500
#endif

#ifndef DNS_CAPTURE_PORT
#define DNS_CAPTURE_PORT 15053
#endif

// 127.0.0.6 (network order)
static const __u32 sidecar_ip = 127 + (6 << 24);
// ::6 (network order)
static const __u32 sidecar_ip6[4] = {0, 0, 0, 6 << 24};
