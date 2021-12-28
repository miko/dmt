package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/miko/dmt/internal"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var detailed bool

func init() {
	infoCmd.Flags().BoolVarP(&detailed, "detailed", "l", false, "Detailed info mode")
	rootCmd.AddCommand(infoCmd)
}

func infoShowDatabaseMetaInfo(data2 internal.DatabaseState) {
	dbversion := data2.CurrentVersion
	fmt.Println("Database info:")
	if dbversion < 0 {
		fmt.Println("Database not initialized!")
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Version", "Index", "Date", "Dgraph"})
		table.Append([]string{fmt.Sprintf("%d", dbversion), data2.IndexLocation, data2.Date.Format(time.RFC3339), viper.GetString("dgraph")})
		table.Render()
	}
}

func infoShowDatabaseInfo(data internal.IndexState, data2 internal.DatabaseState) {
	dbversion := data2.CurrentVersion
	max := len(data.Entries)
	min := len(data2.Entries)
	L := data.Entries
	nextstate := "Pending"
	if len(data2.Entries) > max {
		max = len(data2.Entries)
		min = len(data.Entries)
		L = data2.Entries
		nextstate = "Future"
	}
	if max == 0 {
		fmt.Println("No index file found! Use dmt build to build one, reference it with -i flag.")
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"V#", "State", "Filename", "Executed on", "Check", "Description"})

	for k := 0; k < min; k++ {
		v1 := data.Entries[k]
		v := data2.Entries[k]
		var state, date string
		if k+1 <= dbversion {
			state = "Applied"
			date = data2.Entries[k].Date.Format(time.RFC3339)
		} else {
			state = "Pending"
			date = "-"
		}
		desc := v.Description
		if desc == "" {
			desc = v1.Description
		}
		check := "OK"
		if v.Filename != v1.Filename {
			check = "BAD FILENAME"
		}
		if v.MD5SUM != v1.MD5SUM {
			check = "BAD"
		}
		table.Append([]string{fmt.Sprintf("%d", k+1), state, v.Filename, date, check, desc})
	}

	for k := min; k < max; k++ {
		v := L[k]
		date := "-"
		if &L[k].Date != nil && !L[k].Date.IsZero() {
			date = L[k].Date.Format(time.RFC3339)
		}
		table.Append([]string{fmt.Sprintf("%d", k+1), nextstate, v.Filename, date, "", v.Description})
	}
	table.Render()
}

var infoCmd = &cobra.Command{
	Use:          "info",
	Short:        "Migrations info",
	Long:         `Migrations info`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		data2, err := internal.GetDatabaseState()
		if err != nil {
			fmt.Println("Database not initialized!")
		} else {
			infoShowDatabaseMetaInfo(data2)
		}

		fmt.Println()

		data, err := internal.GetIndexState(data2.IndexLocation)
		if err != nil {
			fmt.Println(err)
		} else {
			err = internal.VerifyIndexState(&data)
			if err != nil {
				fmt.Println(err)
			}
		}

		if detailed {
			//			infoShowDatabaseLongInfo(data, data2)
			infoShowDatabaseInfo(data, data2)
		} else {
			infoShowDatabaseInfo(data, data2)
		}
		return err
	},
}
