package controller

import (
	"github.com/hazelcast/hazelcast-go-operator/pkg/controller/hazelcast"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, hazelcast.Add)
}
