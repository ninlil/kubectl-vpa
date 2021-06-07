package main

import (
	"fmt"
)

type modeArgs struct {
	Mode  modeEnum `arg:"positional,required" help:"What mode to set the VPA in: Off, Initial or Auto" placeholder:"MODE"`
	Names []string `arg:"positional,required" help:"Name(s) of the VPA-resources to modify" placeholder:"NAME"`
}

func (mode *modeArgs) Verify() error {
	if len(mode.Names) == 0 {
		return fmt.Errorf("no names specified")
	}
	return nil
}

func (mode *modeArgs) Exec(k8 *k8client, args *cmdArgs) {
	fmt.Printf("set mode %s\n", args.Mode.Mode)
	for _, input := range args.Mode.Names {
		ns, name := args.getParts(input)

		fmt.Printf("on %s / %s: ", ns, name)
		err := k8.PatchVPA(ns, name, "/spec/updatePolicy/updateMode", args.Mode.Mode.String())
		if err != nil {
			fmt.Printf("error: %s\n", err)
		} else {
			fmt.Println("ok")
		}
	}
}
