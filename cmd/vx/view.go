package main

import "github.com/spf13/cobra"

func newViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <target>",
		Short: "Inspect a local vpkg package, export, or direct .vxt file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
}
