package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HazelcastSpec defines the desired state of Hazelcast
type HazelcastSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Size int32 `json:"size"`

	Config corev1.ConfigMap `json:"config"`

	Service corev1.Service `json:"service"`
}

// HazelcastStatus defines the observed state of Hazelcast
type HazelcastStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
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
