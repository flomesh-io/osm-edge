package multicluster

import (
	multiclusterv1alpha1 "github.com/openservicemesh/osm/pkg/apis/multicluster/v1alpha1"
	"github.com/openservicemesh/osm/pkg/k8s/informers"
	"github.com/openservicemesh/osm/pkg/service"
)

func (c *Client) getGlobalTrafficPolicy(svc service.MeshService) *multiclusterv1alpha1.GlobalTrafficPolicy {
	gblTrafficPolicyIf, exists, err := c.informers.GetByKey(informers.InformerKeyGlobalTrafficPolicy, svc.NamespacedKey())
	if !exists || err != nil {
		return nil
	}

	return gblTrafficPolicyIf.(*multiclusterv1alpha1.GlobalTrafficPolicy)
}

func (c *Client) isLocality(svc service.MeshService) bool {
	gblTrafficPolicy := c.getGlobalTrafficPolicy(svc)
	if gblTrafficPolicy == nil {
		return true
	}
	if gblTrafficPolicy.Spec.LbType == multiclusterv1alpha1.LocalityLbType {
		return true
	}
	return false
}

//func (c *Client) isFailOver(svc service.MeshService) bool {
//	gblTrafficPolicy := c.getGlobalTrafficPolicy(svc)
//	if gblTrafficPolicy == nil {
//		return false
//	}
//	if gblTrafficPolicy.Spec.LbType == multiclusterv1alpha1.FailOverLbType {
//		return true
//	}
//	return false
//}
