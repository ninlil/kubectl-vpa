package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mickep76/encoding"
	_ "github.com/mickep76/encoding/json"
	_ "github.com/mickep76/encoding/toml"
	_ "github.com/mickep76/encoding/yaml"
)

type createArgs struct {
	Names  []string   `arg:"positional,required" help:"Pod-name(s)to create VPA for" placeholder:"NAME"`
	Mode   modeEnum   `arg:"-m,--mode" help:"Assign the VPA mode to the output"`
	Format formatEnum `arg:"-o,--output-format" help:"Select output format (yaml [default], json, toml)"`
}

func (cr *createArgs) Verify() error {
	if len(cr.Names) == 0 {
		return fmt.Errorf("no names specified")
	}
	return nil
}

func (cr *createArgs) Exec(k8 *k8client, args *cmdArgs) {

	if args.Debug {
		fmt.Printf("## format as %s\n", cr.Format.String())
	}

	yaml, err := cr.Format.Encoder()
	if err != nil {
		log.Printf("yaml-encoder-error: %v", err)
		return
	}

	for _, input := range cr.Names {
		ns, name := args.getParts(input)

		switch true {
		case cr.createForPod(k8, ns, name, yaml, args):
		case cr.createForDaemonSet(k8, ns, name, yaml, args):
		case cr.createForStatefulSet(k8, ns, name, yaml, args):
		case cr.createForDeployment(k8, ns, name, yaml, args):
		case cr.createForCronJob(k8, ns, name, yaml, args):
		case cr.createForCronJobBeta(k8, ns, name, yaml, args):
		default:
			fmt.Fprintf(os.Stderr, "error: unable to locate resource %s/%s\n", ns, name)
		}
	}
}

type vpaRoot struct {
	APIVersion string  `yaml:"apiVersion"`
	Kind       string  `yaml:"kind"`
	Meta       vpaMeta `yaml:"metadata"`
	Spec       vpaSpec `yaml:"spec"`
}
type vpaMeta struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}
type vpaSpec struct {
	RsrcPolicy   vpaRsrcPolicy   `yaml:"resourcePolicy"`
	Target       vpaTarget       `yaml:"targetRef"`
	UpdatePolicy vpaUpdatePolicy `yaml:"updatePolicy"`
}
type vpaRsrcPolicy struct {
	Containers []vpaContPolicy `yaml:"containerPolicies"`
}
type vpaContPolicy struct {
	Name     string   `yaml:"containerName"`
	MinAllow vpaAllow `yaml:"minAllowed,omitempty"`
	Mode     string   `yaml:"mode"`
}
type vpaAllow struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}
type vpaTarget struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
}
type vpaUpdatePolicy struct {
	UpdateMode string `yaml:"updateMode"`
}

func (cr *createArgs) createForDaemonSet(k8 *k8client, ns, name string, enc encoding.Codec, args *cmdArgs) bool {

	ds, err := k8.DaemonSet(ns, name)
	if err != nil {
		return false
	}
	var cnames []string

	for _, c := range ds.Spec.Template.Spec.Containers {
		cnames = append(cnames, c.Name)
	}

	cr.createVPA(enc, "DaemonSet", ns, name, cnames, args)
	return true
}

func (cr *createArgs) createForStatefulSet(k8 *k8client, ns, name string, enc encoding.Codec, args *cmdArgs) bool {

	ss, err := k8.StatefulSet(ns, name)
	if err != nil {
		return false
	}
	var cnames []string

	for _, c := range ss.Spec.Template.Spec.Containers {
		cnames = append(cnames, c.Name)
	}

	cr.createVPA(enc, "StatefulSet", ns, name, cnames, args)
	return true
}

func (cr *createArgs) createForDeployment(k8 *k8client, ns, name string, enc encoding.Codec, args *cmdArgs) bool {

	dep, err := k8.Deployment(ns, name)
	if err != nil {
		return false
	}
	var cnames []string

	for _, c := range dep.Spec.Template.Spec.Containers {
		cnames = append(cnames, c.Name)
	}

	cr.createVPA(enc, "Deployment", ns, name, cnames, args)
	return true
}

func (cr *createArgs) createForCronJob(k8 *k8client, ns, name string, enc encoding.Codec, args *cmdArgs) bool {

	job, err := k8.CronJob(ns, name)
	if err != nil {
		return false
	}
	var cnames []string

	for _, c := range job.Spec.JobTemplate.Spec.Template.Spec.Containers {
		cnames = append(cnames, c.Name)
	}

	cr.createVPA(enc, "CronJob", ns, name, cnames, args)
	return true
}

func (cr *createArgs) createForCronJobBeta(k8 *k8client, ns, name string, enc encoding.Codec, args *cmdArgs) bool {

	job, err := k8.CronJobBeta(ns, name)
	if err != nil {
		return false
	}
	var cnames []string

	for _, c := range job.Spec.JobTemplate.Spec.Template.Spec.Containers {
		cnames = append(cnames, c.Name)
	}

	cr.createVPA(enc, "CronJobBeta", ns, name, cnames, args)
	return true
}

func (cr *createArgs) createForPod(k8 *k8client, ns, name string, enc encoding.Codec, args *cmdArgs) bool {

	pod, err := k8.Pod(ns, name)
	if err != nil {
		return false
	}

	if args.Debug {
		fmt.Printf("\n# create vpa-yaml for pod %s/%s\n", pod.Name, pod.Namespace)
	}

	kind := kindPod
	ns = pod.Namespace
	name = pod.Name

	for i, owner := range pod.GetOwnerReferences() {
		kind = owner.Kind
		name = owner.Name
		if args.Debug {
			fmt.Printf("# owner[%d] = %s '%s'\n", i, owner.Kind, owner.Name)
		}
	}

	switch kind {
	case kindReplicaSet:
		kind = kindDeployment
	case kindJob:
		kind = kindCronJob
	}

	var cnames = make([]string, 0, len(pod.Spec.Containers))
	for _, c := range pod.Spec.Containers {
		cnames = append(cnames, c.Name)
	}

	cr.createVPA(enc, kind, ns, name, cnames, args)
	return true
}

func (cr *createArgs) createVPA(enc encoding.Codec, kind, ns, name string, containers []string, args *cmdArgs) {
	var version string

	switch kind {
	case kindReplicaSet, kindDeployment:
		kind = kindDeployment
		version = "apps/v1"
	case kindJob:
		kind = kindCronJob
		version = "batch/v1beta1"
	case kindStatefulSet, kindDaemonSet:
		version = "apps/v1"
	case kindCronJob:
		version = "batch/v1"
	case kindCronJobBeta:
		kind = kindCronJob
		version = "batch/v1beta1"
	default:
		log.Printf("Unsupported kind: %s", kind)
		return
	}

	if version == "" {
		fmt.Printf("# error for %s %s/%s: unable to determine target apiVersion\n", kind, ns, name)
		return
	}

	if args.Debug {
		fmt.Printf("# create for %s@%s %s/%s\n", kind, version, ns, name)
	}

	var vpa = vpaRoot{
		APIVersion: "autoscaling.k8s.io/v1",
		Kind:       kindVPA,
		Meta: vpaMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: vpaSpec{
			RsrcPolicy: vpaRsrcPolicy{
				Containers: []vpaContPolicy{
					{
						Name:     "*",
						Mode:     "Auto",
						MinAllow: vpaAllow{CPU: "10m", Memory: "10Mi"},
					},
				},
			},
			Target: vpaTarget{
				APIVersion: version,
				Kind:       kind,
				Name:       name,
			},
			UpdatePolicy: vpaUpdatePolicy{
				UpdateMode: cr.Mode.String(),
			},
		},
	}

	for _, cname := range containers {
		vpa.Spec.RsrcPolicy.Containers = append(vpa.Spec.RsrcPolicy.Containers,
			vpaContPolicy{
				Name:     cname,
				Mode:     "Auto",
				MinAllow: vpaAllow{CPU: "10m", Memory: "10Mi"},
			})
	}

	buf, err := enc.Encode(vpa)
	if err != nil {
		log.Printf("error encoding for %s %s/%s", kind, ns, name)
		return
	}

	fmt.Println("---")
	fmt.Print(string(buf))
}
