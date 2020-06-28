package hazelcast

import (
	"fmt"
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-go-operator/pkg/apis/hazelcast/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getSecurityContext(cr *hazelcastv1alpha1.Hazelcast) *v1.PodSecurityContext {
	user := HazelcastSecurityContextUser
	var securityContext = v1.PodSecurityContext{
		RunAsUser:    &user,
		RunAsGroup:   &user,
		RunAsNonRoot: boolP(true),
		FSGroup:      &user,
	}
	if cr.Spec.StatefulSet != nil && cr.Spec.StatefulSet.SecurityContext != nil {
		securityContext = *cr.Spec.StatefulSet.SecurityContext
	}
	return &securityContext
}

func boolP(b bool) *bool {
	p := b
	return &p
}

func getReplicas(cr *hazelcastv1alpha1.Hazelcast) *int32 {
	var replicas int32 = 1
	if cr.Spec.StatefulSet == nil {
		return &replicas
	}
	if cr.Spec.StatefulSet.Replicas <= 0 {
		return &replicas
	} else {
		return &cr.Spec.StatefulSet.Replicas
	}
}

func getVolumes(cr *hazelcastv1alpha1.Hazelcast) []v1.Volume {
	var volumes []v1.Volume
	volumes = []v1.Volume{{
		Name: HazelcastConfigName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: HazelcastConfigName,
				},
			},
		},
	}}
	// TODO append custom volume for custom jars
	return volumes
}

func getVolumeMounts(cr *hazelcastv1alpha1.Hazelcast) []v1.VolumeMount {
	var volumeMounts []v1.VolumeMount

	volumeMounts = append(volumeMounts, v1.VolumeMount{
		Name:      HazelcastConfigName,
		MountPath: "/data/hazelcast",
	})
	// TODO append custom volume mounts for custom jars

	return volumeMounts
}

func getProbe(cr *hazelcastv1alpha1.Hazelcast, probeType string, delay, period, timeout, success, failure int32) *v1.Probe {
	var path string
	if probeType == "liveness" {
		path = HazelcastLivenessProbeEndpoint
	} else if probeType == "readiness" {
		path = HazelcastReadinessProbeEndpoint
	}

	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: path,
				Port: intstr.FromInt(int(GetHazelcastPort(cr))),
			},
		},
		InitialDelaySeconds: delay,
		PeriodSeconds:       period,
		TimeoutSeconds:      timeout,
		SuccessThreshold:    success,
		FailureThreshold:    failure,
	}
}

func getJavaOpts(namespace string, maxWaitSeconds int32) string {
	return fmt.Sprintf("-Dhazelcast.config=/data/hazelcast/hazelcast.yaml -DserviceName=%s -Dnamespace=%s -Dhazelcast.shutdownhook.policy=GRACEFUL -Dhazelcast.shutdownhook.enabled=true -Dhazelcast.graceful.shutdown.max.wait=%v", HazelcastServiceName, namespace, maxWaitSeconds)
}

func getContainers(cr *hazelcastv1alpha1.Hazelcast) []v1.Container {
	var containers []v1.Container

	containers = append(containers, v1.Container{
		Name:            "hazelcast",
		Image:           fmt.Sprintf("%s:%s", HazelcastImage, HazelcastVersion),
		ImagePullPolicy: "IfNotPresent",
		Ports: []v1.ContainerPort{
			{
				Name:          "hazelcast",
				ContainerPort: int32(GetHazelcastPort(cr)),
				Protocol:      "TCP",
			},
			// TODO metric port
		},
		Env: []v1.EnvVar{
			{
				Name:  "JAVA_OPTS",
				Value: getJavaOpts(cr.Namespace, 600),
			},
		},
		//Resources:                getResources(cr),
		VolumeMounts:   getVolumeMounts(cr),
		LivenessProbe:  getProbe(cr, "liveness", 30, 10, 10, 1, 10),
		ReadinessProbe: getProbe(cr, "readiness", 30, 10, 10, 1, 10),
	})
	return containers
}

func getStatefulSetSpec(cr *hazelcastv1alpha1.Hazelcast, configHash string) appsv1.StatefulSetSpec {
	return appsv1.StatefulSetSpec{
		Replicas:    getReplicas(cr),
		ServiceName: HazelcastServiceName,
		Selector: &metav1.LabelSelector{
			MatchLabels: labelsForHazelcast(cr),
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      labelsForHazelcast(cr),
				Annotations: map[string]string{"lastConfigHash": configHash},
			},
			Spec: v1.PodSpec{
				SecurityContext:    getSecurityContext(cr),
				Volumes:            getVolumes(cr),
				Containers:         getContainers(cr),
				ServiceAccountName: HazelcastServiceAccountName,
			},
		},
	}
}

func GetHazelcastStatefulSet(cr *hazelcastv1alpha1.Hazelcast, configHash string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      HazelcastStatefulSetName,
			Namespace: cr.Namespace,
		},
		Spec:   getStatefulSetSpec(cr, configHash),
		Status: appsv1.StatefulSetStatus{},
	}
}

func (r *ReconcileHazelcast) statefulsetForHazelcast(cr *hazelcastv1alpha1.Hazelcast, configHash string) *appsv1.StatefulSet {
	statefulset := GetHazelcastStatefulSet(cr, configHash)

	controllerutil.SetControllerReference(cr, statefulset, r.scheme)
	return statefulset
}

func GetHazelcastStatefulsetSelector(cr *hazelcastv1alpha1.Hazelcast) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      HazelcastStatefulSetName,
	}
}
