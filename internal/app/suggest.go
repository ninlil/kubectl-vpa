package app

import (
	"fmt"
	"log"
	"math"
	"strconv"
)

type suggestData struct {
	Resources suggestResources `json:"resources"`
}
type suggestResources struct {
	Requests suggestValues `json:"requests"`
	Limits   suggestValues `json:"limits"`
}
type suggestValues struct {
	CPU    *string `json:"cpu,omitempty"`
	Memory *string `json:"memory,omitempty"`
}

type suggestArgs struct {
	Name   string     `arg:"positional,required" help:"Name of the VPA-resource to create suggestion" placeholder:"NAME"`
	Format formatEnum `arg:"-o,--output-format" help:"Select output format (yaml [default], json, toml)"`
}

var (
	errNameMissing = fmt.Errorf("resource name must be specified")
)

func (suggest *suggestArgs) Verify() error {
	if suggest.Name == "" {
		return errNameMissing
	}
	return nil
}

func (suggest *suggestArgs) Exec(k8 *k8client, args *cmdArgs) {
	ns, name := args.getParts(args.Suggest.Name)
	vpa, err := k8.VPA(ns, name)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	recommend := vpa.Status.Recommendation
	if recommend == nil {
		log.Printf("VPA %s/%s have no recommendations (yet)", vpa.Namespace, vpa.Name)
		return
	}

	yaml, err := suggest.Format.Encoder()
	if err != nil {
		log.Printf("yaml-encoder-error: %v", err)
		return
	}

	for _, c := range recommend.ContainerRecommendations {
		fmt.Printf("\n# container %s\n", c.ContainerName)
		var data suggestData

		if v, ok := c.Target["cpu"]; ok {
			data.Resources.Requests.CPU = calcValue(v.String(), 1)
		}
		if v, ok := c.Target["memory"]; ok {
			data.Resources.Requests.Memory = calcValue(v.String(), 1)
		}
		if v, ok := c.UpperBound["cpu"]; ok {
			data.Resources.Limits.CPU = calcValue(v.String(), 1.5)
		}
		if v, ok := c.UpperBound["memory"]; ok {
			data.Resources.Limits.Memory = calcValue(v.String(), 1.5)
		}

		buf, err := yaml.Encode(&data)
		if err != nil {
			log.Printf("yaml-encoder-error: %v", err)
			return
		}

		fmt.Print(string(buf))
	}

	//fmt.Printf("vpa = %s\n", vpa.Name)
}

const (
	multMi = 1024 * 1024
)

func calcValue(v string, scale float64) *string {
	n, err := getValue(v)
	if err != nil {
		return nil
	}

	var suffix string
	n *= scale

	if n < 10 {
		n *= 1000
		suffix = "m"
	} else {
		if n > multMi {
			n /= multMi
			suffix = "Mi"
		}
	}
	txt := fmt.Sprintf("%d%s", int(math.RoundToEven(n)), suffix)
	return &txt
}

func getValue(v string) (float64, error) {

	var p int = -1
	for i, ch := range v {
		if !((ch >= '0' && ch <= '9') || (ch == '.')) && p < 0 {
			p = i
		}
	}
	if p < 0 {
		p = len(v)
	}

	var suffix string
	n, err := strconv.ParseFloat(v[:p], 64)
	if p >= 0 {
		suffix = v[p:]
	}

	switch suffix {
	case "m":
		n /= 1000
	case "Mi":
		n *= multMi
	}

	return n, err
}
