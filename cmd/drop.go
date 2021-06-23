package cmd

import (
	"fmt"

	"github.com/miko/dmt/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(dropCmd)
}

var dropCmd = &cobra.Command{
	Use:          "drop",
	Short:        "Drop data",
	Long:         `Drop data`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		vrb := viper.GetBool("verbose")

		if viper.GetString("export_url") != "" {
			data2, err := internal.GetDatabaseState()
			if err != nil {
				fmt.Println(err)
				return
			}
			if data2.CurrentVersion > 0 {
				if vrb {
					fmt.Printf("[warn] Exporting all data to %s before dropping data - version was at %d", viper.GetString("export_url"), data2.CurrentVersion)
				}
				_, err = internal.ExportData()
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				if vrb {
					fmt.Printf("[warn] NOT Exporting all data to %s before dropping data - version was at %d", viper.GetString("export_url"), data2.CurrentVersion)
				}
			}
		}

		err := internal.DropAll()
		if err != nil {
			//			log.Fatal(err)
			fmt.Println(err)
			return
		}
		fmt.Println("Database dropped. You should initialize it now.")
	},
}
