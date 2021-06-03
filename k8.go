package main

import (
	"context"
	"encoding/json"
	"os"

	vpa "github.com/ninlil/kubectl-vpa/vpa_v1"
	// v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	vpaCRD = "verticalpodautoscalers"
)

type k8client struct {
	k8Client  *kubernetes.Clientset
	vpaClient *rest.RESTClient
}

//  patchStringValue specifies a patch operation for a string.
type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

//  patchUint32Value specifies a patch operation for a uint32.
type patchUInt32Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value uint32 `json:"value"`
}

func connect(args *CmdArgs) (*k8client, error) {

	_ = vpa.AddToScheme(scheme.Scheme)

	// uses the current context in kubeconfig
	// path-to-kubeconfig -- for example, /root/.kube/config
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &vpa.SchemeGroupVersion //schema.GroupVersion{Group: v1alpha1.GroupName, Version: v1alpha1.GroupVersion}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	exampleRestClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		return nil, err
	}

	return &k8client{
		k8Client:  clientset,
		vpaClient: exampleRestClient,
	}, nil
}

func (k8 *k8client) Pods(ns string) v1.PodInterface {
	return k8.k8Client.CoreV1().Pods(ns)
}

func (k8 *k8client) VPAs(ns string) (*vpa.VerticalPodAutoscalerList, error) {
	result := vpa.VerticalPodAutoscalerList{}
	var req = k8.vpaClient.Get().Resource(vpaCRD)
	if ns != "" {
		req = req.Namespace(ns)
	}
	err := req.Do(context.Background()).Into(&result)
	return &result, err
}

func (k8 *k8client) VPA(ns, name string) (*vpa.VerticalPodAutoscaler, error) {
	result := vpa.VerticalPodAutoscaler{}
	err := k8.vpaClient.Get().Resource(vpaCRD).Namespace(ns).Name(name).Do(context.Background()).Into(&result)
	return &result, err
}

func (k8 *k8client) PatchString(ns, name, path, value string) error {
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  path,
		Value: value,
	}}
	payloadBytes, _ := json.Marshal(payload)
	result := k8.vpaClient.Patch(types.JSONPatchType).Resource(vpaCRD).Namespace(ns).Name(name).Body(payloadBytes).Do(context.Background())
	if err := result.Error(); err != nil {
		return err
	}

	return nil
}
