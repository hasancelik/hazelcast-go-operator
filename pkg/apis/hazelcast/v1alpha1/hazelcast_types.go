package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HazelcastSpec defines the desired state of Hazelcast
type HazelcastSpec struct {
	Size        int32                 `json:"size"`
	HostPort    int32                 `json:"hostPort"`
	StatefulSet *HazelcastStatefulSet `json:"statefulset,omitempty"`
	Config      *HazelcastConfig      `json:"config"`
	Service     *HazelcastService     `json:"service"`
}

type HazelcastService struct {
	Name      string           `json:"name,omitempty"`
	Type      v1.ServiceType   `json:"type,omitempty"`
	Ports     []v1.ServicePort `json:"ports,omitempty"`
	ClusterIP string           `json:"clusterIP,omitempty"`
}

type HazelcastConfig struct {
	Name string            `json:"name,omitempty"`
	Data map[string]string `json:"data",omitempty`
}

type HazelcastStatefulSet struct {
	Annotations     map[string]string      `json:"annotations,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	Replicas        int32                  `json:"replicas"`
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	//NodeSelector    map[string]string      `json:"nodeSelector,omitempty"`
	//Tolerations     []v1.Toleration        `json:"tolerations,omitempty"`
	//Affinity        *v1.Affinity           `json:"affinity,omitempty"`
}

// HazelcastStatus defines the observed state of Hazelcast
type HazelcastStatus struct {
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Hazelcast is the Schema for the hazelcasts API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=hazelcasts,scope=Namespaced
type Hazelcast struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HazelcastSpec   `json:"spec,omitempty"`
	Status HazelcastStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HazelcastList contains a list of Hazelcast
type HazelcastList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hazelcast `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hazelcast{}, &HazelcastList{})
}
