package control

import (
	"github.com/spf13/cobra"
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
)

func settings() []harness.Setting[config.Config] {
	return harness.Registered[config.Config](&component.Settings)
}

func setCommand() *cobra.Command {
	return harness.Says(&cobra.Command{
		Use:   "settings [name] [value...]",
		Short: "List the device's settings, read one, or change one",
	}, set)
}

func set(args []string) (string, error) { return harness.Settings(config.Get, settings(), args) }
