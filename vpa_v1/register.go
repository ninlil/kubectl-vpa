package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is the version of the current VPA-scheme
var SchemeGroupVersion = schema.GroupVersion{Group: "autoscaling.k8s.io", Version: "v1"}

// Global vars to enable VPA to use the k8-restclient
var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&VerticalPodAutoscaler{},
		&VerticalPodAutoscalerList{},
		// &VerticalPodAutoscalerCheckpoint{},
		// &VerticalPodAutoscalerCheckpointList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	// metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
