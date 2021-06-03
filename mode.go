package main

import (
	"fmt"
	"strings"
)

func mode(k8 *k8client, args *CmdArgs) {

	fmt.Printf("set mode %s\n", args.Mode.Mode)
	for _, name := range args.Mode.Names {
		var ns = args.Namespace
		parts := strings.SplitN(name, "/", 2)
		if len(parts) > 1 {
			ns = parts[0]
			name = parts[1]
		}
		if ns == "" {
			ns = "default"
		}

		fmt.Printf("on %s / %s\n", ns, name)
	}
}
