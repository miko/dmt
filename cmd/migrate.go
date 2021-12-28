package cmd

import (
	"fmt"
	"os"

	"github.com/miko/dmt/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	from        int
	to          int
	successFile string
	header      string
	forceDrop   bool
)

func init() {
	migrateCmd.Flags().IntVarP(&from, "from", "f", 0, "From target")
	migrateCmd.Flags().IntVarP(&to, "to", "t", 0, "To target")
	migrateCmd.Flags().StringVarP(&successFile, "file", "s", "", "Touch this file in case of successful migration")
	migrateCmd.Flags().BoolVarP(&forceDrop, "force-drop", "x", false, "Force drop all database data before migration")
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:          "migrate",
	Short:        "Migrate database to newer version",
	Long:         `Migrate database to newer version`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		vrb := viper.GetBool("verbose")
		if vrb {
			fmt.Printf("[info] Started migrator %s\n", VER)
		}

		idx := viper.GetString("index")
		if idx == "" {
			idx = "_dmt.json"
		}

		data2, err := internal.GetDatabaseState()
		if err != nil {
			fmt.Println(err)
			return err
		}

		if viper.GetString("export_url") != "" {
			if viper.GetString("export_url") != "none" {
				if data2.CurrentVersion > 0 {
					if vrb {
						fmt.Printf("[warn] Exporting all data to %s before migration - version was at %d\n", viper.GetString("export_url"), data2.CurrentVersion)
					}
					_, err = internal.ExportData()
					if err != nil {
						fmt.Println(err)
						return err
					}
				}
			} else {
				fmt.Println("Skipping data export before migration.")
			}
		}
		if forceDrop {
			if vrb {
				fmt.Printf("[warn] Dropping all data before migration - version was at %d\n", data2.CurrentVersion)
			}
			err = internal.DropAll()
			if err != nil {
				fmt.Println(err)
				return err
			}
			// FIXME TODO - wait for db response

			data2, err = internal.GetDatabaseState()
			if err != nil {
				fmt.Println(err)
				return err
			}
		}

		if data2.CurrentVersion < 0 {
			if vrb {
				fmt.Printf("[warn] Database was not initialized, initializing now\n")
			}
			err = internal.InitializeDatabase(idx)
			if err != nil {
				fmt.Println(err)
				return err
			}
			data2, err = internal.GetDatabaseState()
			if err != nil {
				fmt.Println(err)
				return err
			}
			if data2.CurrentVersion < 0 {
				return fmt.Errorf("Could not initialize database\n")
			}
			if vrb {
				fmt.Printf("[info] Database initialized with index %s\n", idx)
			}
		}
		data, err := internal.GetIndexState(data2.IndexLocation)
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = internal.VerifyIndexState(&data)
		if err != nil {
			fmt.Println(err)
			return err
		}
		if from == 0 {
			from = data2.CurrentVersion
		}
		if from > 0 && from != data2.CurrentVersion {
			fmt.Printf("Cannot migrate from %d, database at version %d\n", from, data2.CurrentVersion)
			return fmt.Errorf("Cannot migrate from %d, database at version %d\n", from, data2.CurrentVersion)
		}
		if to == 0 {
			to = len(data.Entries)
		}
		if to == from {
			fmt.Printf("Did not migrate, database already at version %d\n", data2.CurrentVersion)
			return nil
		}
		if to < from {
			fmt.Printf("Cannot migrate to lower version %d - not yet supported, database now at version %d\n", to, data2.CurrentVersion)
			return fmt.Errorf("Cannot migrate to lower version %d - not yet supported, database now at version %d\n", to, data2.CurrentVersion)
		}
		if to > len(data.Entries) {
			fmt.Printf("Cannot migrate to version %d - know about %d versions\n", to, len(data.Entries))
			return fmt.Errorf("Cannot migrate to version %d - know about %d versions\n", to, len(data.Entries))
		}
		for k := from + 1; k <= to; k++ {
			err = internal.UpVersion(k, data.Entries[k-1])
			if err != nil {
				fmt.Printf("Cannot migrate: %s\n", err)
				return fmt.Errorf("Cannot migrate: %s\n", err)
			}
		}
		if successFile != "" {
			f, err := os.Create(successFile)
			if err != nil {
				fmt.Printf("[error] cannot create file %s\n", successFile)
				return err
			}
			f.Close()
		}
		return nil
	},
}
