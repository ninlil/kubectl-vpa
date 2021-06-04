package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
)

type Subcommand interface {
	Verify() error
	Exec(*k8client, *CmdArgs)
}

type CmdArgs struct {
	Namespace     string       `arg:"-n,--namespace" help:"namespace to compare" default:"default"`
	AllNamespaces bool         `arg:"-A,--all-namespaces" help:"If present, list the requested object(s) across all namespaces."`
	Debug         bool         `arg:"-d,--debug" help:"enable debug output"`
	Kubeconfig    string       `arg:"-k" help:"filename of kubeconfig to use"`
	Compare       *CompareArgs `arg:"subcommand:compare" help:"Compare pod requests to VPA recommendations"`
	Mode          *ModeArgs    `arg:"subcommand:mode" help:"Change mode on VPA-resource(s)"`
	Suggest       *SuggestArgs `arg:"subcommand:suggest" help:"Suggest YAML from a VPA-resource"`
}

type compareFilter struct {
	filter      bool
	showOff     bool
	showInitial bool
	showAuto    bool
}

type ModeEnum int

const (
	modeOff ModeEnum = iota + 1
	modeInitial
	modeAuto
)

func (mode ModeEnum) String() string {
	switch mode {
	case modeInitial:
		return "Initial"
	case modeAuto:
		return "Auto"
	default:
		return "Off"
	}
}

func (mode *ModeEnum) UnmarshalText(b []byte) error {
	s := strings.ToLower(string(b))
	switch s {
	case "off":
		*mode = modeOff
	case "initial":
		*mode = modeInitial
	case "auto":
		*mode = modeAuto
	default:
		return fmt.Errorf("unknown mode: '%s', allowed values: Off, Initial & Auto", mode)
	}
	return nil
}

func (CmdArgs) Version() string {
	return "vpa 0.4.0"
}

func parseArgs() (Subcommand, *CmdArgs, bool) {
	var args CmdArgs
	pa := arg.MustParse(&args)
	if pa == nil {
		log.Println("unable to parse arguments")
		os.Exit(1)
	}

	if args.AllNamespaces {
		args.Namespace = ""
	}

	if pa.Subcommand() == nil {
		pa.Fail("Command not specified")
	}

	var cmd Subcommand
	if ver, ok := pa.Subcommand().(Subcommand); ok {
		if err := ver.Verify(); err != nil {
			pa.Fail(err.Error())
		}
		cmd = ver
	}

	if cmd == nil {
		pa.Fail("command not implemented")
	}

	return cmd, &args, true
}

func (args *CmdArgs) getParts(input string) (ns, name string) {
	ns = args.Namespace
	parts := strings.SplitN(input, "/", 2)
	if len(parts) > 1 {
		ns = parts[0]
		name = parts[1]
	} else {
		name = input
	}
	if ns == "" {
		ns = "default"
	}
	return ns, name
}
