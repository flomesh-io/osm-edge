// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"net"
	"unsafe"
)

// IP2Pointer returns the pointer of a ip string
func IP2Pointer(ipstr string) (unsafe.Pointer, error) {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return nil, fmt.Errorf("error parse ip: %s", ipstr)
	}
	if ip.To4() != nil {
		// ipv4, we need to clear the bytes
		for i := 0; i < 12; i++ {
			ip[i] = 0
		}
	}
	return unsafe.Pointer(&ip[0]), nil
}
