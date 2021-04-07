package cmd

import (
	"fmt"

	"github.com/miko/dmt/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	from int
	to   int
)

func init() {
	migrateCmd.Flags().IntVarP(&from, "from", "f", 0, "From target")
	migrateCmd.Flags().IntVarP(&to, "to", "t", 0, "To target")
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Print the migrate number of Hugo",
	Long:  `All software has migrates. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		data, err := internal.GetIndexState()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = internal.VerifyIndexState(&data)
		if err != nil {
			fmt.Println(err)
			return
		}
		data2, err := internal.GetDatabaseState()
		if err != nil {
			fmt.Println(err)
			err = internal.InitializeDatabase(viper.GetString("index"), data.BaseDir)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		if from == 0 {
			from = data2.CurrentVersion
		}
		if from > 0 && from != data2.CurrentVersion {
			fmt.Printf("Cannot migrate from %d, database at version %d\n", from, data2.CurrentVersion)
			return
		}
		if to == 0 {
			to = len(data.Entries)
		}
		if to == from {
			fmt.Printf("Cannot migrate, database already at version %d\n", data2.CurrentVersion)
			return
		}
		if to < from {
			fmt.Printf("Cannot migrate to lower version %d - not yet supported, database now at version %d\n", to, data2.CurrentVersion)
			return
		}
		if to > len(data.Entries) {
			fmt.Printf("Cannot migrate to version %d - know about %d versions\n", to, len(data.Entries))
			return
		}
		for k := from + 1; k <= to; k++ {
			err = internal.UpVersion(k, data.Entries[k-1])
			if err != nil {
				fmt.Printf("Cannot migrate: %s\n", err)
				return
			}
		}
	},
}
