package hazelcast

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-go-operator/pkg/apis/hazelcast/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileHazelcast) ensureStatefulSet(
	h *hazelcastv1alpha1.Hazelcast,
	statefulSet *appsv1.StatefulSet,
) (*reconcile.Result, error) {
	// Check if the StatefulSet already exists, if not create a new one
	found := &appsv1.StatefulSet{}
	selector := GetHazelcastStatefulsetSelector(h)
	err := r.client.Get(context.TODO(),  selector, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the StatefulSet
		log.Info("Creating a new StatefulSet", "StatefulSet.Namespace", statefulSet.Namespace, "StatefulSet.Name", statefulSet.Name)
		err = r.client.Create(context.TODO(), statefulSet)

		if err != nil {
			// StatefulSet failed
			log.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", statefulSet.Namespace, "StatefulSet.Name", statefulSet.Name)
			return &reconcile.Result{}, err
		} else {
			// StatefulSet was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the StatefulSet not existing
		log.Error(err, "Failed to get StatefulSet")
		return &reconcile.Result{}, err
	}
	return nil, nil
}

func (r *ReconcileHazelcast) checkStatefulSetConfigHash(
	statefulSet *appsv1.StatefulSet,
	configHash string,
) (*reconcile.Result, error) {
	currentHash := statefulSet.Spec.Template.Annotations["lastConfigHash"]
	if !strings.EqualFold(configHash, currentHash) {
		statefulSetCopy := statefulSet.DeepCopy()
		statefulSetCopy.Spec.Template.Annotations["lastConfigHash"] = configHash
		patch := client.MergeFrom(statefulSet)
		patchErr := r.client.Patch(context.TODO(), statefulSetCopy, patch)
		if patchErr != nil {
			log.Error(patchErr, "Failed to Patch Hazelcast StatefulSet")
			return &reconcile.Result{}, patchErr
		}
		log.Info("StatefulSet patched!")
		return &reconcile.Result{Requeue: true}, nil
	}
	return nil, nil
}

func (r *ReconcileHazelcast) checkStatefulSize(h *hazelcastv1alpha1.Hazelcast, statefulSet *appsv1.StatefulSet) (*reconcile.Result, error){
	// Ensure the deployment size is the same as the spec
	size := h.Spec.Size
	if *statefulSet.Spec.Replicas != size {
		statefulSetCopy := statefulSet.DeepCopy()
		*statefulSetCopy.Spec.Replicas = size
		patch := client.MergeFrom(statefulSet)
		patchErr := r.client.Patch(context.TODO(), statefulSetCopy, patch)
		if patchErr != nil {
			log.Error(patchErr, "Failed to Patch Hazelcast StatefulSet Replica count")
			return &reconcile.Result{}, patchErr
		}

		// Spec updated - return and requeue
		return &reconcile.Result{Requeue: true}, nil
	}
	return nil, nil
}

func (r *ReconcileHazelcast) ensureService(
	h *hazelcastv1alpha1.Hazelcast,
	service *corev1.Service,
) (*reconcile.Result, error) {
	found := &corev1.Service{}
	selector := GetHazelcastServiceSelector(h)
	err := r.client.Get(context.TODO(), selector, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		log.Info("Creating a new Service for Hazelcast", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return &reconcile.Result{}, err
		}  else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get Service")
		return &reconcile.Result{}, err
	}
	return nil, nil
}

func (r *ReconcileHazelcast) ensureConfigMap(
	h *hazelcastv1alpha1.Hazelcast,
	configMap *corev1.ConfigMap,
) (*reconcile.Result, error) {
	found := &corev1.ConfigMap{}
	selector := GetHazelcastConfigMapSelector(h)
	err := r.client.Get(context.TODO(), selector, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the configMap
		log.Info("Creating a new ConfigMap for Hazelcast", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
		err = r.client.Create(context.TODO(), configMap)
		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
			return &reconcile.Result{}, err
		}  else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the configMap not existing
		log.Error(err, "Failed to get ConfigMap")
		return &reconcile.Result{}, err
	} else {
		// ConfigMap is already created so check its hash
		config := h.Spec.Config.Data[HazelcastConfigFileName]
		configHash := generateSHA1CheckSum(&config)
		existingConfig := found.Data[HazelcastConfigFileName]
		existingConfigHash := generateSHA1CheckSum(&existingConfig)
		if !strings.EqualFold(existingConfigHash, configHash) {
			found.Data[HazelcastConfigFileName] = config
			if len(found.ObjectMeta.Annotations) == 0 {
				found.ObjectMeta.Annotations = map[string]string{"lastConfigHash": configHash}
			} else {
				found.ObjectMeta.Annotations["lastConfigHash"] = configHash
			}
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				log.Error(err, "Failed to update Hazelcast Pod(s)' configuration(s)!")
				return &reconcile.Result{}, err
			}
			return &reconcile.Result{Requeue: true}, nil
		}
	}
	return nil, nil
}

func labelsForHazelcast(h *hazelcastv1alpha1.Hazelcast) map[string]string {
	return map[string]string{
		"app": "hazelcast",
		"hazelcast_cr": h.Name,
	}
}

func generateSHA1CheckSum(str *string) string {
	if str != nil {
		sha1 := sha1.New()
		sha1.Write([]byte(*str))
		checkSum := sha1.Sum(nil)
		return hex.EncodeToString(checkSum)
	}
	return ""
}