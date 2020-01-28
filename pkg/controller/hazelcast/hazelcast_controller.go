package hazelcast

import (
	"context"
	"reflect"
	"strings"

	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-go-operator/pkg/apis/hazelcast/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_hazelcast")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Hazelcast Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileHazelcast{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("hazelcast-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Hazelcast
	err = c.Watch(&source.Kind{Type: &hazelcastv1alpha1.Hazelcast{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Hazelcast
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hazelcastv1alpha1.Hazelcast{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileHazelcast implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileHazelcast{}

// ReconcileHazelcast reconciles a Hazelcast object
type ReconcileHazelcast struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Hazelcast object and makes changes based on the state read
// and what is in the Hazelcast.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileHazelcast) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Hazelcast")

	// Fetch the Hazelcast instance
	hazelcast := &hazelcastv1alpha1.Hazelcast{}
	err := r.client.Get(context.TODO(), request.NamespacedName, hazelcast)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Hazelcast resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Hazelcast")
		return reconcile.Result{}, err
	}

	foundConfigMap := &corev1.ConfigMap{}
	configMapName := hazelcast.Spec.Config.ObjectMeta.Name
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: configMapName, Namespace: hazelcast.Namespace}, foundConfigMap)
	if err != nil && errors.IsNotFound(err) {
		configMap := r.configMapForHazelcast(hazelcast)
		reqLogger.Info("Creating a new ConfigMap for Hazelcast", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
		err = r.client.Create(context.TODO(), configMap)
		if err != nil {
			reqLogger.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get ConfigMap")
		return reconcile.Result{}, err
	}

	// Check if the StatefulSet already exists, if not create a new one
	found := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: hazelcast.Name, Namespace: hazelcast.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new statefulSet
		statefulSet := r.statefulSetForHazelcast(hazelcast)
		reqLogger.Info("Creating a new StatefulSet", "StatefulSet.Namespace", statefulSet.Namespace, "StatefulSet.Name", statefulSet.Name)
		err = r.client.Create(context.TODO(), statefulSet)
		if err != nil {
			reqLogger.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", statefulSet.Namespace, "StatefulSet.Name", statefulSet.Name)
			return reconcile.Result{}, err
		}
		// StatefulSet created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get StatefulSet")
		return reconcile.Result{}, err
	}

	foundService := &corev1.Service{}
	serviceSpec := hazelcast.Spec.Service
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: serviceSpec.ObjectMeta.Name, Namespace: hazelcast.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		service := r.serviceForHazelcast(hazelcast)
		reqLogger.Info("Creating a new Service for Hazelcast", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := hazelcast.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update StatefulSet", "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	configYAML := hazelcast.Spec.Config.Data["hazelcast.yaml"]
	existingConfigYAML := foundConfigMap.Data["hazelcast.yaml"]
	if !strings.EqualFold(existingConfigYAML, configYAML) {
		foundConfigMap.Data["hazelcast.yaml"] = configYAML
		if len(foundConfigMap.ObjectMeta.Labels) == 0 {
			foundConfigMap.ObjectMeta.Labels = map[string]string{"configMap": "updated"}
		} else {
			foundConfigMap.ObjectMeta.Labels["configMap"] = "updated"
		}
		err = r.client.Update(context.TODO(), foundConfigMap)
		if err != nil {
			reqLogger.Error(err, "Failed to update Hazelcast Pod(s) configuration(s)")
			return reconcile.Result{}, err
		}
		reqLogger.Info("ConfigMap is updated!")
		return reconcile.Result{Requeue: true}, nil
	}

	if len(foundConfigMap.ObjectMeta.Labels) != 0 {
		if _, ok := foundConfigMap.Labels["configMap"]; ok {
			statefulSet := &appsv1.StatefulSet{}
			if err := r.client.Get(context.TODO(), types.NamespacedName{Name: hazelcast.Name, Namespace: hazelcast.Namespace}, statefulSet); err != nil {
				if !errors.IsNotFound(err) {
					return reconcile.Result{}, err
				}
			}
			statefulSetCopy := statefulSet.DeepCopy()
			statefulSetCopy.Spec.Template.Annotations = map[string]string{"configMap": "updated"}
			patch := client.MergeFrom(statefulSet)
			patchErr := r.client.Patch(context.TODO(), statefulSetCopy, patch)
			if patchErr != nil {
				reqLogger.Error(patchErr, "Failed to Patch Hazelcast StatefulSet")
				return reconcile.Result{}, patchErr
			}
			reqLogger.Info("StatefulSet is patched!")
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// Update the Hazelcast status with the pod names
	// List the pods for this hazelcast's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(hazelcast.Namespace),
		client.MatchingLabels(labelsForHazelcast(hazelcast.Name)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Hazelcast.Namespace", hazelcast.Namespace, "Hazelcast.Name", hazelcast.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, hazelcast.Status.Nodes) {
		hazelcast.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), hazelcast)
		if err != nil {
			reqLogger.Error(err, "Failed to update Hazelcast status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// stateFullSetForHazelcast returns a hazelcast StatefullSet object
func (r *ReconcileHazelcast) statefulSetForHazelcast(hazelcast *hazelcastv1alpha1.Hazelcast) *appsv1.StatefulSet {
	ls := labelsForHazelcast(hazelcast.Name)
	replicas := hazelcast.Spec.Size
	serviceName := hazelcast.Spec.Service.ObjectMeta.Name
	configMapName := hazelcast.Spec.Config.ObjectMeta.Name

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hazelcast.Name,
			Namespace: hazelcast.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: serviceName,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "hazelcast/hazelcast:3.12.5",
						Name:  "hazelcast",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 5701,
							Name:          "hazelcast",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "hazelcast-storage",
							MountPath: "/data/hazelcast",
						}},
						Env: []corev1.EnvVar{{
							Name:  "JAVA_OPTS",
							Value: "-Dhazelcast.rest.enabled=true -Dhazelcast.config=/data/hazelcast/hazelcast.yaml",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "hazelcast-storage",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: configMapName,
								},
							},
						},
					}},
				},
			},
		},
	}

	// Set Hazelcast instance as the owner and controller
	controllerutil.SetControllerReference(hazelcast, statefulSet, r.scheme)
	return statefulSet
}

// serviceForHazelcast returns a service object
func (r *ReconcileHazelcast) serviceForHazelcast(hazelcast *hazelcastv1alpha1.Hazelcast) *corev1.Service {
	serviceSpec := hazelcast.Spec.Service.Spec
	serviceMetadata := hazelcast.Spec.Service.ObjectMeta

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceMetadata.Name,
			Namespace: hazelcast.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     serviceSpec.Type,
			Selector: map[string]string{"app": "hazelcast", "hazelcast_cr": hazelcast.Name},
			Ports:    serviceSpec.Ports,
		},
	}

	controllerutil.SetControllerReference(hazelcast, service, r.scheme)
	return service
}

func (r *ReconcileHazelcast) configMapForHazelcast(hazelcast *hazelcastv1alpha1.Hazelcast) *corev1.ConfigMap {
	configMapName := hazelcast.Spec.Config.ObjectMeta.Name
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: hazelcast.Namespace,
		},
		Data: hazelcast.Spec.Config.Data,
	}

	controllerutil.SetControllerReference(hazelcast, configMap, r.scheme)
	return configMap
}

// labelsForHazelcast returns the labels for selecting the resources
// belonging to the given hazelcast CR name.
func labelsForHazelcast(name string) map[string]string {
	return map[string]string{"app": "hazelcast", "hazelcast_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
