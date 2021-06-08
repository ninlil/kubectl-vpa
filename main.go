package main

import (
	"fmt"
	"os"

	"github.com/ninlil/kubectl-vpa/internal/app"
)

// types from https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go

func main() {
	cmd, args, ok := app.ParseArgs()
	if !ok {
		return
	}

	k8, err := app.Connect(args)
	if err != nil {
		fmt.Printf("kubernetes-error: %v", err)
		os.Exit(1)
	}

	cmd.Exec(k8, args)
}
