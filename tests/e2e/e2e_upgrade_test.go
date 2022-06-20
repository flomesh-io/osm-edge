package e2e

import (
	. "github.com/openservicemesh/osm/tests/framework"
)

var _ = OSMDescribe("Upgrade from latest",
	OSMDescribeInfo{
		Tier:   2,
		Bucket: 10,
	},
	func() {

	})
