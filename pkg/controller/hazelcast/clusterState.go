package hazelcast

import (
	"context"
	"github.com/hazelcast/hazelcast-go-operator/pkg/apis/hazelcast/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterState struct {
	HazelcastStatefulSet    *appsv1.StatefulSet
	HazelcastService        *v1.Service
	HazelcastConfig         *v1.ConfigMap
}

func NewClusterState() *ClusterState {
	return &ClusterState{}
}

func (i *ClusterState) Read(ctx context.Context, cr *v1alpha1.Hazelcast, client client.Client) error {

	err := i.readHazelcastConfig(ctx, cr, client)
	if err != nil {
		return err
	}

	err = i.readHazelcastService(ctx, cr, client)
	if err != nil {
		return err
	}

	err = i.readHazelcastStatefulSet(ctx, cr, client)
	if err != nil {
		return err
	}

	return nil
}


func (i *ClusterState) readHazelcastConfig(ctx context.Context, cr *v1alpha1.Hazelcast, client client.Client) error {
	currentState, err := GetHazelcastConfigMap(cr)
	if err != nil {
		return err
	}
	selector := GetHazelcastConfigMapSelector(cr)
	err = client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.HazelcastConfig = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readHazelcastService(ctx context.Context, cr *v1alpha1.Hazelcast, client client.Client) error {
	currentState := GetHazelcastService(cr)
	selector := GetHazelcastServiceSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.HazelcastService = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readHazelcastStatefulSet(ctx context.Context, cr *v1alpha1.Hazelcast, client client.Client) error {
	currentState := GetHazelcastStatefulSet(cr, "")
	selector := GetHazelcastStatefulsetSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.HazelcastStatefulSet = currentState.DeepCopy()
	return nil
}