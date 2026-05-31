package main

import "github.com/spf13/cobra"

func newGenCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "gen <target>",
		Aliases: []string{"generate"},
		Short:   "Preview or apply generation for a local vpkg export or .vxt file",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
}
