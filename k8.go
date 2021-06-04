package main

import (
	"context"
	"encoding/json"
	"os"

	// v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	vpa "github.com/ninlil/kubectl-vpa/vpa_v1"
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

func connect(args *CmdArgs) (*k8client, error) {

	_ = vpa.AddToScheme(scheme.Scheme)

	// uses the current context in kubeconfig
	// path-to-kubeconfig -- for example, /root/.kube/config
	config, err := clientcmd.BuildConfigFromFlags(args.Kubeconfig, os.Getenv("KUBECONFIG"))
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

func (k8 *k8client) Pods(ns string) typev1.PodInterface {
	return k8.k8Client.CoreV1().Pods(ns)
}

func (k8 *k8client) Pod(ns, name string) (*corev1.Pod, error) {
	return k8.k8Client.CoreV1().Pods(ns).Get(context.Background(), name, metav1.GetOptions{})
}

func (k8 *k8client) DaemonSet(ns, name string) (*appsv1.DaemonSet, error) {
	return k8.k8Client.AppsV1().DaemonSets(ns).Get(context.Background(), name, metav1.GetOptions{})
}

func (k8 *k8client) StatefulSet(ns, name string) (*appsv1.StatefulSet, error) {
	return k8.k8Client.AppsV1().StatefulSets(ns).Get(context.Background(), name, metav1.GetOptions{})
}

func (k8 *k8client) Deployment(ns, name string) (*appsv1.Deployment, error) {
	return k8.k8Client.AppsV1().Deployments(ns).Get(context.Background(), name, metav1.GetOptions{})
}

func (k8 *k8client) CronJob(ns, name string) (*batchv1.CronJob, error) {
	return k8.k8Client.BatchV1().CronJobs(ns).Get(context.Background(), name, metav1.GetOptions{})
}

func (k8 *k8client) CronJobBeta(ns, name string) (*batchv1beta1.CronJob, error) {
	return k8.k8Client.BatchV1beta1().CronJobs(ns).Get(context.Background(), name, metav1.GetOptions{})
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

// PatchVPA
//
// Adapted from example: https://gist.github.com/dwmkerr/447692c8bba28929ef914239781c4e59
func (k8 *k8client) PatchVPA(ns, name, path, value string) error {
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
