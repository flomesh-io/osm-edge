package injector

import (
	corev1 "k8s.io/api/core/v1"
)

// GetVolumeSpec returns a list of volumes to add to the POD
func GetVolumeSpec(sidecarBootstrapConfigName string) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: sidecarBootstrapConfigVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: sidecarBootstrapConfigName,
				},
			},
		},
	}
}
