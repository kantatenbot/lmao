package command

import (
	"github.com/kantatenbot/mass-exec/internal/version"
	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:     "mass-exec",
	Short:   "xargs in ThE cLoUd",
	Version: version.Version,
}

func Execute() error {
	return rootCommand.Execute()
}
