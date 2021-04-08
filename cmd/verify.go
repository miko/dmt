package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(verifyCmd)
}

var verifyCmd = &cobra.Command{
	Use:          "verify",
	Short:        "Verify migrations",
	Long:         `Verify migrations`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Not implemented")
		return fmt.Errorf("Not implemented")
	},
}
