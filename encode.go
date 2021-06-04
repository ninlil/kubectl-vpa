package main

import (
	"fmt"
	"strings"

	"github.com/mickep76/encoding"
	_ "github.com/mickep76/encoding/json"
	_ "github.com/mickep76/encoding/toml"
	_ "github.com/mickep76/encoding/yaml"
)

type FormatEnum int

const (
	formatYAML FormatEnum = iota
	formatJSON
	formatTOML
)

func (f *FormatEnum) UnmarshalText(b []byte) error {
	s := strings.ToLower(string(b))
	switch s {
	case "yaml":
		*f = formatYAML
	case "json":
		*f = formatJSON
	case "toml":
		*f = formatTOML
	default:
		return fmt.Errorf("unknown mode: '%s', allowed values: yaml, json & toml", s)
	}
	return nil
}

func (f FormatEnum) String() string {
	switch f {
	case formatYAML:
		return "yaml"
	case formatJSON:
		return "json"
	case formatTOML:
		return "toml"
	}
	return "yaml"
}

func (f FormatEnum) Encoder() (encoding.Codec, error) {
	switch f {
	case formatYAML:
		return encoding.NewCodec(f.String(), encoding.WithMapString())
	case formatJSON:
		return encoding.NewCodec(f.String(), encoding.WithIndent("  "))
	default:
		return encoding.NewCodec(f.String())
	}
}
