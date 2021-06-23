package cmd

import (
	"fmt"

	"github.com/miko/dmt/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:          "export",
	Short:        "Export data",
	Long:         `Export data to s3 endpoint`,
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
					fmt.Printf("[info] Exporting all data - version was at %d\n", data2.CurrentVersion)
				}
				_, err = internal.ExportData()
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("[info] Database exported.")
			} else {
				if vrb {
					fmt.Printf("[warn] NOT Exporting all data to %s - version was at %d\n", viper.GetString("export_url"), data2.CurrentVersion)
				}
			}
		} else {
			fmt.Println("[warn] Database not exported - no URL given")
		}

	},
}
