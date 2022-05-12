package app

import (
	"fmt"
	"log"
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
)

type subcommand interface {
	Verify() error
	Exec(*k8client, *cmdArgs)
}

type cmdArgs struct {
	Namespace     string       `arg:"-n,--namespace" help:"namespace to compare" default:"default"`
	AllNamespaces bool         `arg:"-A,--all-namespaces" help:"If present, list the requested object(s) across all namespaces."`
	Debug         bool         `arg:"-d,--debug" help:"enable debug output"`
	Kubeconfig    string       `arg:"-k" help:"filename of kubeconfig to use"`
	Compare       *compareArgs `arg:"subcommand:compare" help:"Compare pod requests to VPA recommendations"`
	Mode          *modeArgs    `arg:"subcommand:mode" help:"Change mode on VPA-resource(s)"`
	Suggest       *suggestArgs `arg:"subcommand:suggest" help:"Suggest YAML from a VPA-resource"`
	Create        *createArgs  `arg:"subcommand:create" help:"Create a VPA-YAML from a pod"`
}

type modeEnum int

const (
	modeOff modeEnum = iota + 1
	modeInitial
	modeAuto

	modeOffText     = "Off"
	modeInitialText = "Initial"
	modeAutoText    = "Auto"
)

func (mode modeEnum) String() string {
	switch mode {
	case modeInitial:
		return modeInitialText
	case modeAuto:
		return modeAutoText
	default:
		return modeOffText
	}
}

func (mode *modeEnum) UnmarshalText(b []byte) error {
	s := strings.ToLower(string(b))
	switch s {
	case "off":
		*mode = modeOff
	case "initial", "init":
		*mode = modeInitial
	case "auto":
		*mode = modeAuto
	default:
		return fmt.Errorf("unknown mode: '%s', allowed values: Off, Initial & Auto", mode)
	}
	return nil
}

func (cmdArgs) Version() string {
	return "vpa 0.7.2"
}

func ParseArgs() (subcommand, *cmdArgs, bool) {
	var args cmdArgs
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

	var cmd subcommand
	if ver, ok := pa.Subcommand().(subcommand); ok {
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

func (args *cmdArgs) getParts(input string) (ns, name string) {
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
