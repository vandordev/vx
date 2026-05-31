package main

import (
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	pkg "github.com/vandordev/vx/internal/package"
)

// Metadata loaded from package.toml at build time
var (
	version = pkg.Version()
	name    = pkg.Name()
	short   = pkg.Short()
)

type rootOptions struct {
	configPath  string
	showVersion bool
}

var rootCmd = newRootCmd()

// Execute is the CLI entrypoint.
func Execute() error {
	return rootCmd.Execute()
}

func newRootCmd() *cobra.Command {
	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use:   name,
		Short: short,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.showVersion {
				cmd.Printf("%s\n", resolvedVersion())
				return nil
			}
			return runOverview(cmd)
		},
	}

	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "", "config file path")
	cmd.Flags().BoolVarP(&opts.showVersion, "version", "v", false, "print version information")

	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newCompletionCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newGenCmd())

	return cmd
}

func resolvedVersion() string {
	ver := version
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ver
	}
	if ver == "dev" && strings.TrimSpace(info.Main.Version) != "" && info.Main.Version != "(devel)" {
		ver = info.Main.Version
	}
	return ver
}

func runOverview(cmd *cobra.Command) error {
	lines := []string{
		"VX inspects and generates local templates from a project vpkg/ runtime.",
		"",
		"Commands:",
		"  vx view <target>      Inspect a package, export, or direct .vxt template",
		"  vx gen <target>       Preview or apply generation for a template target",
		"",
		"Examples:",
		"  vx view vandor/go-backend-core:usecase",
		"  vx gen vandor/go-backend-core:default",
		"  vx view ./templates/usecase.vxt",
		"  vx gen ./templates/usecase.vxt",
		"",
		"vx expects the current project, or one of its parents, to contain vpkg/.",
	}

	cmd.Println(strings.Join(lines, "\n"))
	return nil
}
