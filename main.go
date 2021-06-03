package main

import (
	"fmt"
	"os"
)

// types from https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go

func main() {
	args, ok := parseArgs()
	if !ok {
		return
	}

	k8, err := connect(args)
	if err != nil {
		fmt.Printf("kubernetes-error: %v", err)
		os.Exit(1)
	}

	switch true {
	case args.Compare != nil:
		compare(k8, args)
	case args.Mode != nil:
		mode(k8, args)
	case args.Suggest != nil:
		suggest(k8, args)
	}
}
