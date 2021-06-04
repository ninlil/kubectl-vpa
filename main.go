package main

import (
	"fmt"
	"os"
)

// types from https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go

func main() {
	cmd, args, ok := parseArgs()
	if !ok {
		return
	}

	k8, err := connect(args)
	if err != nil {
		fmt.Printf("kubernetes-error: %v", err)
		os.Exit(1)
	}

	cmd.Exec(k8, args)
}
