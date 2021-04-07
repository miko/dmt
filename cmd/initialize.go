package cmd

import (
	"fmt"

	"github.com/miko/dmt/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initializeCmd)
}

var initializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Initialize database",
	Long:  `Initialize database`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.GetIndexState()
		data, err := internal.GetIndexState()
		if err != nil {
			//			log.Fatal(err)
			fmt.Println(err)
			return
		}
		err = internal.InitializeDatabase(data.IndexFile, data.BaseDir)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Database initialized")
	},
}
