package cmd

import (
	"fmt"

	"github.com/miko/dmt/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dropCmd)
}

var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Drop data",
	Long:  `Drop data`,
	Run: func(cmd *cobra.Command, args []string) {
		err := internal.DropAll()
		if err != nil {
			//			log.Fatal(err)
			fmt.Println(err)
			return
		}
		fmt.Println("Database dropped. You should initialize it now.")
	},
}
