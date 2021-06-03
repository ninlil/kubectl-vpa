package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
)

type CompareArgs struct {
	AllPods bool          `arg:"-0,--no-vpa" help:"all pods, even those without a VPA"`
	Modes   []string      `arg:"-m,--mode,separate" help:"filter only VPAs with specified mode(s)" placeholder:"MODE"`
	Head    int           `arg:"-h,--head" help:"only print N first lines" default:"-1"`
	Tail    int           `arg:"-t,--tail" help:"only print N last lines" default:"-1"`
	Sort    []int         `arg:"-s,--sort,separate" help:"sort by column N (negative sorts descending)"`
	filter  compareFilter `arg:"-"`
}

type ModeArgs struct {
	Mode  string   `arg:"-m,--mode" help:"What mode to set the VPA in: Off, Initial or Auto" placeholder:"MODE"`
	Names []string `arg:"positional" placeholder:"NAME"`
}

type CmdArgs struct {
	Namespace     string       `arg:"-n,--namespace" help:"namespace to compare" default:"default"`
	AllNamespaces bool         `arg:"-A,--all-namespaces" help:"If present, list the requested object(s) across all namespaces."`
	Debug         bool         `arg:"-d,--debug" help:"enable debug output"`
	Compare       *CompareArgs `arg:"subcommand:compare"`
	Mode          *ModeArgs    `arg:"subcommand:mode"`
}

type compareFilter struct {
	filter      bool
	showOff     bool
	showInitial bool
	showAuto    bool
}

func (CmdArgs) Version() string {
	return "vpa 0.2.0"
}

func parseArgs() (*CmdArgs, bool) {
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

	if ver, ok := pa.Subcommand().(Subcommand); ok {
		if err := ver.Verify(&args); err != nil {
			pa.Fail(err.Error())
		}
	}

	return &args, true
}

type Subcommand interface {
	Verify(*CmdArgs) error
}

func (comp *CompareArgs) Verify(args *CmdArgs) error {
	for _, mode := range comp.Modes {
		switch strings.ToLower(mode) {
		case "off":
			comp.filter.filter = true
			comp.filter.showOff = true
		case "initial":
			comp.filter.filter = true
			comp.filter.showInitial = true
		case "auto":
			comp.filter.filter = true
			comp.filter.showAuto = true
		default:
			return fmt.Errorf("unknown mode: '%s', allowed values: Off, Initial & Auto", mode)
		}
	}
	return nil
}

func (mode *ModeArgs) Verify(args *CmdArgs) error {
	switch strings.ToLower(mode.Mode) {
	case "off", "initial", "auto":
	default:
		return fmt.Errorf("unknown mode: '%s', allowed values: Off, Initial & Auto", mode)
	}
	if len(mode.Names) == 0 {
		return fmt.Errorf("no names specified")
	}
	return nil
}
