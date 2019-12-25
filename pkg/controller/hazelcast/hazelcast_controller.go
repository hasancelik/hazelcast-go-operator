package hazelcast

import (
	"context"
	"reflect"

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

	// Update status.Nodes if neededhazelcast
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

// deploymentForHazelcast returns a hazelcast Deployment object
func (r *ReconcileHazelcast) statefulSetForHazelcast(hazelcast *hazelcastv1alpha1.Hazelcast) *appsv1.StatefulSet {
	ls := labelsForHazelcast(hazelcast.Name)
	replicas := hazelcast.Spec.Size

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hazelcast.Name,
			Namespace: hazelcast.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
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
					}},
				},
			},
		},
	}

	// Set Memcached instance as the owner and controller
	controllerutil.SetControllerReference(hazelcast, statefulSet, r.scheme)
	return statefulSet
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
