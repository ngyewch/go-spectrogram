package main

import (
	"github.com/ngyewch/go-spectrogram/build_info"
	"github.com/ngyewch/go-spectrogram/cmd"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

func main() {
	build_info.BuildInfo = build_info.BuildConfig{
		Version: version,
		Commit:  commit,
		Date:    date,
		BuiltBy: builtBy,
	}

	cmd.Execute()
}
