package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(verifyCmd)
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify migrations",
	Long:  `Verify migrations`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Not implemented")
		return
	},
}
