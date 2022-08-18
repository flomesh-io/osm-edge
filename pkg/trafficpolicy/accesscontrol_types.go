package trafficpolicy

// AccessControlTrafficPolicy defines the access control traffic match and routes for a given backend
type AccessControlTrafficPolicy struct {
	TrafficMatches    []*AccessControlTrafficMatch
	HTTPRoutePolicies []*InboundTrafficPolicy
}

// AccessControlTrafficMatch defines the attributes to match access control traffic for a given backend
type AccessControlTrafficMatch struct {
	Name                     string
	Port                     uint32
	Protocol                 string
	SourceIPRanges           []string
	ServerNames              []string
	SkipClientCertValidation bool
}
