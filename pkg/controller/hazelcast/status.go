package hazelcast

import (
	"context"
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-go-operator/pkg/apis/hazelcast/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileHazelcast) updateCRStatus(h *hazelcastv1alpha1.Hazelcast) (*reconcile.Result, error) {
	// Update the Hazelcast status with the pod names
	// List the pods for this hazelcast's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(h.Namespace),
		client.MatchingLabels(labelsForHazelcast(h)),
	}
	if err := r.client.List(context.TODO(), podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Hazelcast.Namespace", h.Namespace, "Hazelcast.Name", h.Name)
		return &reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, h.Status.Nodes) {
		h.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), h)
		if err != nil {
			log.Error(err, "Failed to update Hazelcast status")
			return &reconcile.Result{}, err
		}
	}
	return nil, nil
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
