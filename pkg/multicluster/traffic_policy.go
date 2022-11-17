package multicluster

import (
	multiclusterv1alpha1 "github.com/openservicemesh/osm/pkg/apis/multicluster/v1alpha1"
	"github.com/openservicemesh/osm/pkg/constants"
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
	if gblTrafficPolicy != nil {
		if gblTrafficPolicy.Spec.LbType == multiclusterv1alpha1.LocalityLbType {
			return true
		}
		return false
	}
	return true
}

func (c *Client) getServiceTrafficPolicy(svc service.MeshService) (lbType multiclusterv1alpha1.LoadBalancerType, clusterKeys map[string]int) {
	gblTrafficPolicy := c.getGlobalTrafficPolicy(svc)
	if gblTrafficPolicy != nil {
		lbType = gblTrafficPolicy.Spec.LbType
		if len(gblTrafficPolicy.Spec.LoadBalanceTarget) > 0 {
			clusterKeys = make(map[string]int)
			for _, lbt := range gblTrafficPolicy.Spec.LoadBalanceTarget {
				clusterKeys[lbt.ClusterKey] = lbt.Weight
			}
		}
		return
	}
	lbType = multiclusterv1alpha1.LocalityLbType
	return
}

// GetLbWeightForService retrieves load balancer type and weight for service
func (c *Client) GetLbWeightForService(svc service.MeshService) (aa, fo, lc bool, weight int) {
	gblTrafficPolicy := c.getGlobalTrafficPolicy(svc)
	if gblTrafficPolicy != nil {
		if gblTrafficPolicy.Spec.LbType == multiclusterv1alpha1.ActiveActiveLbType {
			aa = true
			if len(gblTrafficPolicy.Spec.LoadBalanceTarget) == 0 {
				weight = constants.ClusterWeightAcceptAll
			} else {
				for _, lbt := range gblTrafficPolicy.Spec.LoadBalanceTarget {
					weight += lbt.Weight
					break
				}
			}
			return
		}
		if gblTrafficPolicy.Spec.LbType == multiclusterv1alpha1.FailOverLbType {
			fo = true
			weight = constants.ClusterWeightFailOver
			return
		}
	}
	lc = true
	return
}
