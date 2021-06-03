package main

import (
	"fmt"
	"log"
)

func mode(k8 *k8client, args *CmdArgs) {

	fmt.Printf("set mode %s\n", args.Mode.Mode)
	for _, input := range args.Mode.Names {
		ns, name := args.getParts(input)

		fmt.Printf("on %s / %s\n", ns, name)
		err := k8.PatchString(ns, name, "/spec/updatePolicy/updateMode", args.Mode.Mode)
		if err != nil {
			log.Printf("patch-error: %s", err)
			return
		}
	}
}
