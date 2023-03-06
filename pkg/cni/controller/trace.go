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

package controller

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"unsafe"

	"github.com/openservicemesh/osm/pkg/cni/controller/helpers"
)

const debug = false

// IPAddr supports IPv6
type IPAddr [4]byte

func (ip *IPAddr) String() string {
	return net.IPv4(ip[0], ip[1], ip[2], ip[3]).String()
}

// PodConfig is the config of meshed pod
type PodConfig struct {
	StatusPort       uint16
	Pad              uint16             // pad `json:"-"`
	ExcludeOutRanges [maxItemLen]cidr   `json:"-"`
	IncludeOutRanges [maxItemLen]cidr   `json:"-"`
	IncludeInPorts   [maxItemLen]uint16 `json:"-"`
	IncludeOutPorts  [maxItemLen]uint16 `json:"-"`
	ExcludeInPorts   [maxItemLen]uint16 `json:"-"`
	ExcludeOutPorts  [maxItemLen]uint16 `json:"-"`
}

func tracePodFibMap() {
	if !debug {
		return
	}
	localPodIpsMap := helpers.GetPodFibMap()
	entries := localPodIpsMap.Iterate()
	key := make([]byte, 16)
	value := PodConfig{}
	fmt.Println("-----------[osm_pod_fib]-----------")
	for entries.Next(unsafe.Pointer(&key[0]), unsafe.Pointer(&value)) {
		ipv4 := net.IPv4(key[12], key[13], key[14], key[15])
		bytes, _ := json.Marshal(value)
		fmt.Println("GetPodFibMap.Iterate:", ipv4.String(), string(bytes))
	}
}

// Pair 4 tuples
type Pair struct {
	srcIP   []byte
	dstIP   []byte
	srcPort uint16
	dstPort uint16
}

func (p *Pair) String() string {
	srcIP := net.IPv4(p.srcIP[12], p.srcIP[13], p.srcIP[14], p.srcIP[15])
	dstIP := net.IPv4(p.dstIP[12], p.dstIP[13], p.dstIP[14], p.dstIP[15])
	return fmt.Sprintf(`Pair{ %s:%d->%s:%d }`, srcIP.String(), p.srcPort, dstIP.String(), p.dstPort)
}

// OriginInfo is a wrapper for the original request information.
type OriginInfo struct {
	dstIP   []byte
	procID  uint32
	dstPort uint16
	// last bit means that ip of process is detected.
	flags uint16
}

func (g *OriginInfo) String() string {
	ip := net.IPv4(g.dstIP[12], g.dstIP[13], g.dstIP[14], g.dstIP[15])
	return fmt.Sprintf(`Origin{ Proc[%d] DstIP[%s] DstPort[%d] Flags[%d] }`, g.procID, ip.String(), g.dstPort, g.flags)
}

func traceNatFibMap() {
	if !debug {
		return
	}
	pairOriginalMap := helpers.GetNatFibMap()
	entries := pairOriginalMap.Iterate()
	key := make([]byte, 36)
	value := make([]byte, 24)

	pair := Pair{}
	pair.srcIP = make([]byte, 16)
	pair.dstIP = make([]byte, 16)

	origin := OriginInfo{}
	origin.dstIP = make([]byte, 16)

	fmt.Println("-----------[osm_nat_fib]-----------")
	for entries.Next(unsafe.Pointer(&key[0]), unsafe.Pointer(&value[0])) {
		copy(pair.srcIP, key[0:16])
		copy(pair.dstIP, key[16:32])
		pair.srcPort = binary.BigEndian.Uint16(key[32:34])
		pair.dstPort = binary.BigEndian.Uint16(key[34:36])

		copy(origin.dstIP, value[0:16])
		origin.procID = binary.BigEndian.Uint32(value[16:20])
		origin.dstPort = binary.BigEndian.Uint16(value[20:22])
		origin.flags = binary.BigEndian.Uint16(value[22:24])

		fmt.Println("GetNatFibMap.Iterate:", pair.String(), origin.String())
	}
}
