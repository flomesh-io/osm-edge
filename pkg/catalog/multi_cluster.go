package catalog

import (
	mapset "github.com/deckarep/golang-set"

	"github.com/openservicemesh/osm/pkg/service"
	"github.com/openservicemesh/osm/pkg/trafficpolicy"
)

// GetExportTrafficPolicy returns the export policy for the given mesh service
func (mc *MeshCatalog) GetExportTrafficPolicy(svc service.MeshService) (*trafficpolicy.ServiceExportTrafficPolicy, error) {
	exportedRule, err := mc.multiclusterController.GetExportedRule(svc)
	if err != nil {
		return nil, err
	}
	if exportedRule == nil {
		return nil, nil
	}

	exportTrafficPolicy := new(trafficpolicy.ServiceExportTrafficPolicy)

	trafficMatch := &trafficpolicy.ServiceExportTrafficMatch{
		Name:     service.ExportedServiceTrafficMatchName(svc.Name, svc.Namespace, uint16(exportedRule.PortNumber), svc.Protocol),
		Port:     uint32(exportedRule.PortNumber),
		Protocol: svc.Protocol,
	}

	controllerServices := mc.multiclusterController.GetIngressControllerServices()
	if len(controllerServices) > 0 {
		sourceIPSet := mapset.NewSet() // Used to avoid duplicate IP ranges
		for _, controllerService := range controllerServices {
			if endpoints := mc.listEndpointsForService(controllerService); len(endpoints) > 0 {
				for _, ep := range endpoints {
					sourceCIDR := ep.IP.String() + singeIPPrefixLen
					if sourceIPSet.Add(sourceCIDR) {
						trafficMatch.SourceIPRanges = append(trafficMatch.SourceIPRanges, sourceCIDR)
					}
				}
			}
		}
	}

	exportTrafficPolicy.TrafficMatches = append(exportTrafficPolicy.TrafficMatches, trafficMatch)

	return exportTrafficPolicy, nil
}
