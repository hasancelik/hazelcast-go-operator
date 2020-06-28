package hazelcast

import (
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-go-operator/pkg/apis/hazelcast/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func serviceNameFromSpec(cr *hazelcastv1alpha1.Hazelcast) string {
	if len(cr.Spec.Service.Name) == 0 {
		return HazelcastServiceName
	}
	return cr.Spec.Service.Name
}

func getServiceType(cr *hazelcastv1alpha1.Hazelcast) v1.ServiceType {
	if cr.Spec.Service == nil {
		return v1.ServiceTypeClusterIP
	}
	if cr.Spec.Service.Type == "" {
		return v1.ServiceTypeClusterIP
	}
	return cr.Spec.Service.Type
}

func GetHazelcastPort(cr *hazelcastv1alpha1.Hazelcast) int32 {
	if cr.Spec.HostPort == 0 {
		return HazelcastPort
	}

	return cr.Spec.HostPort
}

func getServicePorts(cr *hazelcastv1alpha1.Hazelcast) []v1.ServicePort{
	defaultPorts := []v1.ServicePort{
		{
			Name:       HazelcastPortName,
			Protocol:   "TCP",
			Port:       GetHazelcastPort(cr),
			TargetPort: intstr.FromString(HazelcastPortName),
		},
	}

	if cr.Spec.Service == nil {
		return defaultPorts
	}

	if cr.Spec.Service.Ports == nil {
		return defaultPorts
	}

	return defaultPorts
}

func getClusterIP(cr *hazelcastv1alpha1.Hazelcast) string {
	var headlessService = "None"

	if cr.Spec.Service == nil {
		return headlessService
	}

	if len(cr.Spec.Service.ClusterIP) == 0 {
		return headlessService
	}

	return cr.Spec.Service.ClusterIP
}

func GetHazelcastService(cr *hazelcastv1alpha1.Hazelcast) *v1.Service{
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceNameFromSpec(cr),
			Namespace: cr.Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports:     getServicePorts(cr),
			Selector:  labelsForHazelcast(cr),
			ClusterIP: getClusterIP(cr),
			Type:      getServiceType(cr),
		},
	}
}

func (r *ReconcileHazelcast) serviceForHazelcast(cr *hazelcastv1alpha1.Hazelcast) *v1.Service {
	service := GetHazelcastService(cr)

	controllerutil.SetControllerReference(cr, service, r.scheme)
	return service
}

func GetHazelcastServiceSelector(cr *hazelcastv1alpha1.Hazelcast) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      HazelcastServiceName,
	}
}