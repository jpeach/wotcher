package cli

import "github.com/spf13/cobra"

// NewWatcher ...
func NewWatcher() *cobra.Command {
	cmd := cobra.Command{
		Use:   "wotcher",
		Short: "Watch stuff change in Kubernetes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return &cmd
}
