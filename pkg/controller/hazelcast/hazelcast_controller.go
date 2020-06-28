package hazelcast

import (
	"context"
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-go-operator/pkg/apis/hazelcast/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hazelcastv1alpha1.Hazelcast{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
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
	h := &hazelcastv1alpha1.Hazelcast{}
	err := r.client.Get(context.TODO(), request.NamespacedName, h)
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

	var result *reconcile.Result


	result, err = r.ensureConfigMap(h, r.configMapForHazelcast(h))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureService(h, r.serviceForHazelcast(h))
	if result != nil {
		return *result, err
	}

	currentState := NewClusterState()
	err = currentState.Read(context.TODO(), h, r.client)
	if err != nil {
		log.Error(err, "Error reading state")
		return *result, err
	}
	currentConfigHash := currentState.HazelcastConfig.Annotations["lastConfigHash"]
	result, err = r.ensureStatefulSet(h, r.statefulsetForHazelcast(h,currentConfigHash))
	if result != nil {
		return *result, err
	}

	currentState = NewClusterState()
	err = currentState.Read(context.TODO(), h, r.client)
	if err != nil {
		log.Error(err, "Error reading state")
		return *result, err
	}
	currentConfigHash = currentState.HazelcastConfig.Annotations["lastConfigHash"]
	currentStatefulSet := currentState.HazelcastStatefulSet
	result, err = r.checkStatefulSetConfigHash(currentStatefulSet, currentConfigHash)
	if result != nil {
		return *result, err
	}

	currentState = NewClusterState()
	err = currentState.Read(context.TODO(), h, r.client)
	if err != nil {
		log.Error(err, "Error reading state")
		return *result, err
	}
	currentStatefulSet = currentState.HazelcastStatefulSet
	result, err = r.checkStatefulSize(h, currentStatefulSet)
	if result != nil {
		return *result, err
	}

	result, err = r.updateCRStatus(h)
	if result != nil {
		return *result, err
	}

	return reconcile.Result{}, nil
}
