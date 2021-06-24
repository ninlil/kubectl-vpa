package app

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ninlil/ansi"
	"github.com/ninlil/columns"
)

type compareArgs struct {
	AllPods      bool          `arg:"-l,--all-pods" help:"all pods, even those without a VPA"`
	Modes        []modeEnum    `arg:"-m,--mode,separate" help:"filter only VPAs with specified mode(s)" placeholder:"MODE"`
	InvertFilter bool          `arg:"-!,--invert" help:"invert the mode-filter"`
	Brief        bool          `arg:"-b,--brief" help:"Show in brief format (namespace/vpa_name)"`
	Head         int           `arg:"-H,--head" help:"only print N first lines" default:"-1"`
	Tail         int           `arg:"-t,--tail" help:"only print N last lines" default:"-1"`
	Sort         []int         `arg:"-s,--sort,separate" help:"sort by column N (negative sorts descending)"`
	Sum          bool          `arg:"-z,--sum" help:"add sums to relevant value columns"`
	filter       compareFilter `arg:"-"`
}

type compareFilter struct {
	filter      bool
	showOff     bool
	showInitial bool
	showAuto    bool
}

func (comp *compareArgs) Verify() error {
	for _, mode := range comp.Modes {
		switch mode {
		case modeOff:
			comp.filter.filter = true
			comp.filter.showOff = true
		case modeInitial:
			comp.filter.filter = true
			comp.filter.showInitial = true
		case modeAuto:
			comp.filter.filter = true
			comp.filter.showAuto = true
		}
	}
	if comp.InvertFilter && !comp.filter.filter {
		comp.filter.filter = true
		comp.filter.showOff = true
		comp.filter.showInitial = true
		comp.filter.showAuto = true
	}
	return nil
}

func (comp *compareArgs) Exec(k8 *k8client, args *cmdArgs) {
	pods, err := k8.Pods(args.Namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	result, err := k8.VPAs(args.Namespace)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("There are %d vpas in the cluster\n", len(result.Items))

	vpas := make(map[string]*vpaData)
	for _, v := range result.Items {
		target := v.Spec.TargetRef
		recommend := v.Status.Recommendation
		if target != nil {
			var vpadata = &vpaData{
				namespace:  v.Namespace,
				api:        target.APIVersion,
				kind:       strings.ToLower(target.Kind),
				name:       target.Name,
				containers: make(map[string]*vpaContainerData),
			}
			if v.Spec.UpdatePolicy != nil && v.Spec.UpdatePolicy.UpdateMode != nil {
				vpadata.mode = string(*v.Spec.UpdatePolicy.UpdateMode)
			}

			if recommend != nil {
				for _, values := range recommend.ContainerRecommendations {
					var cont = &vpaContainerData{}
					if v, ok := values.Target["cpu"]; ok {
						cont.cpu = getCPU(&v)
					}
					if v, ok := values.Target["memory"]; ok {
						cont.memory = getMemory(&v)
					}
					vpadata.containers[values.ContainerName] = cont
				}
			}
			vpas[vpadata.Key()] = vpadata
			if args.Debug {
				fmt.Printf("adding vpa %s with %d containers\n", vpadata.Key(), len(vpadata.containers))
			}
			//} else {
			//fmt.Printf("vpa %s/%s have no target\n", v.Namespace, v.Name)
		}
	}

	var podList []podData
	for _, p := range pods.Items {
		if p.Status.Phase == corev1.PodRunning {
			var pod = podData{
				name:       p.Name,
				namespace:  p.Namespace,
				containers: make(map[string]*containerData),
			}
			for _, owner := range p.GetOwnerReferences() {
				pod.ownerAPI = owner.APIVersion
				pod.ownerKind = strings.ToLower(owner.Kind)
				pod.ownerName = owner.Name
			}
			if pod.ownerKind == "job" {
				pod.ownerKind = "cronjob"
				if i := strings.LastIndex(pod.ownerName, "-"); i > 0 {
					pod.ownerName = pod.ownerName[:i]
				}
			}
			if pod.ownerKind == "replicaset" {
				pod.ownerKind = "deployment"
				if i := strings.LastIndex(pod.ownerName, "-"); i > 0 {
					pod.ownerName = pod.ownerName[:i]
				}
			}
			pod.vpa = vpas[pod.Key()]
			for _, c := range p.Spec.Containers {
				cont := &containerData{
					cpu:    getCPU(c.Resources.Requests.Cpu()),
					memory: getMemory(c.Resources.Requests.Memory()),
				}
				if pod.vpa != nil {
					cont.vpa = pod.vpa.containers[c.Name]
				}
				pod.containers[c.Name] = cont
			}
			podList = append(podList, pod)
			if args.Debug {
				fmt.Printf("adding pod %s '%s' with %d containers\n", pod.name, pod.Key(), len(pod.containers))
			}
		}
	}

	var cw *columns.Writer
	if !comp.Brief {
		cw = columns.New(os.Stdout, "< < < < > > > > > > >")
		cw.Headers("Namespace", "Name", "Mode", "Container", "Req-CPU", "VPA-CPU", "CPU diff%", "Req-RAM", "VPA-RAM", "Mem. diff%", "sum(Δ)")
		cw.HeaderSeparator = true
		if comp.Sum {
			cw.Footer(5, columns.Sum(0))
			cw.Footer(6, columns.Sum(0))
			cw.Footer(8, columns.Sum(0))
			cw.Footer(9, columns.Sum(0))
		}
	}

	diffStyle := columns.NewStyle().Suffix("%").ColorFunc(colorDiff)

	var haveVPA bool
	var printed map[string]bool
	for _, pod := range podList {

		for cname, c := range pod.containers {
			var cols = make([]interface{}, 0, 11)
			cols = append(cols, pod.namespace, pod.name)
			if args.Debug {
				fmt.Printf("adding pod %s/%s with container %s to output\n", pod.namespace, pod.name, cname)
			}

			haveVPA = pod.vpa != nil && c.vpa != nil

			if pod.vpa != nil {
				if c.vpa != nil {
					diffCPU := (c.cpu - c.vpa.cpu) * 100 / c.vpa.cpu
					diffMemory := (c.memory - c.vpa.memory) * 100 / c.vpa.memory
					dCPU := columns.Cell(diffCPU).Style(diffStyle)
					dMemory := columns.Cell(diffMemory).Style(diffStyle)

					cols = append(cols, pod.vpa.mode, cname, c.cpu, c.vpa.cpu, dCPU, mem2mb(c.memory), mem2mb(c.vpa.memory), dMemory, diffCPU+diffMemory)
					//fmt.Printf("    %s  %d/%d  %d/%d\n", cname, c.cpu, c.vpa.cpu, c.memory, c.vpa.memory)
				} else {
					cols = append(cols, pod.vpa.mode, cname, c.cpu, nil, nil, mem2mb(c.memory), nil, nil)
					//fmt.Printf("    %s  %d/?  %d/?\n", cname, c.cpu, c.memory)
				}
			} else {
				cols = append(cols, "---", cname, c.cpu, nil, nil, mem2mb(c.memory), nil, nil)
			}
			if args.Compare.AllPods || haveVPA || args.Compare.InvertFilter {
				show := false
				if args.Compare.filter.filter {
					if pod.vpa != nil {
						if args.Compare.filter.showOff && pod.vpa.mode == modeOffText {
							show = true
						}
						if args.Compare.filter.showInitial && pod.vpa.mode == modeInitialText {
							show = true
						}
						if args.Compare.filter.showAuto && pod.vpa.mode == modeAutoText {
							show = true
						}
					}
				} else {
					show = true
				}
				if show != args.Compare.InvertFilter {
					var brief string
					if comp.Brief {
						switch true {
						case pod.vpa != nil:
							brief = fmt.Sprintf("%s/%s", pod.namespace, pod.vpa.name)
						case pod.ownerName != "":
							brief = fmt.Sprintf("%s/%s", pod.namespace, pod.ownerName)
						default:
							brief = fmt.Sprintf("%s/%s", pod.namespace, pod.name)
						}
						if printed == nil {
							printed = make(map[string]bool)
						}
						if !printed[brief] {
							fmt.Println(brief)
							printed[brief] = true
						}
					} else {
						cw.Write(cols...)
					}
				}
			}
		}
	}

	if !comp.Brief {
		if args.Compare.Head >= 0 {
			cw.Head(args.Compare.Head)
		}
		if args.Compare.Tail >= 0 {
			cw.Tail(args.Compare.Tail)
		}
		if len(args.Compare.Sort) > 0 {
			cw.Sort(args.Compare.Sort...)
		} else {
			cw.Sort(1, 2, 4)
		}
		cw.Flush()
	}
}

func colorDiff(o interface{}) (ansi.Style, bool) {
	switch v := o.(type) {
	case float64:
		if v > 10 {
			return ansi.Blue, true
		}
		if v < -10 {
			return ansi.Red, true
		}
	}
	return ansi.Default, false
}

type vpaData struct {
	api        string
	kind       string
	namespace  string
	name       string
	mode       string
	containers map[string]*vpaContainerData
}

type vpaContainerData struct {
	cpu    int64
	memory int64
}

type podData struct {
	name       string
	namespace  string
	ownerAPI   string
	ownerKind  string
	ownerName  string
	vpa        *vpaData
	containers map[string]*containerData
}

type containerData struct {
	vpa    *vpaContainerData
	cpu    int64
	memory int64
}

func getCPU(v *resource.Quantity) int64 {
	if v == nil || v.IsZero() {
		return 0
	}
	return v.MilliValue()
}

func getMemory(v *resource.Quantity) int64 {
	if v == nil || v.IsZero() {
		return 0
	}
	return v.Value()
}

func (v *vpaData) Key() string {
	return fmt.Sprintf("%s:%s@%s/%s", v.api, v.kind, v.namespace, v.name)
}

func (v *podData) Key() string {
	return fmt.Sprintf("%s:%s@%s/%s", v.ownerAPI, v.ownerKind, v.namespace, v.ownerName)
}

func mem2mb(v int64) *columns.CellData {
	return columns.Cell(math.Round(float64(v*10)/1048576) / 10)
}
