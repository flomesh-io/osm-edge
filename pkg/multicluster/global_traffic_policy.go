package multicluster

import (
	multiclusterv1alpha1 "github.com/openservicemesh/osm/pkg/apis/multicluster/v1alpha1"
	"github.com/openservicemesh/osm/pkg/k8s/informers"
	"github.com/openservicemesh/osm/pkg/service"
)

// GetGlobalTrafficPolicies retrieves global traffic policies
func (c *Client) GetGlobalTrafficPolicies(svc service.MeshService) (*multiclusterv1alpha1.GlobalTrafficPolicy, error) {
	gblTrafficPolicyIf, exists, err := c.informers.GetByKey(informers.InformerKeyGlobalTrafficPolicy, svc.NamespacedKey())
	if !exists || err != nil {
		return nil, err
	}

	gblTrafficPolicy := gblTrafficPolicyIf.(*multiclusterv1alpha1.GlobalTrafficPolicy)

	return gblTrafficPolicy, nil
}
