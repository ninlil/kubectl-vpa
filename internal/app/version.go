package app

import (
	_ "embed"
)

//go:generate bash ../../get_version.sh
//go:embed version.txt
var version string

func versionFunc() string {
	return version
}
