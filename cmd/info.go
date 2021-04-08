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
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Version", "Index", "BaseDir", "Date", "Dgraph"})
	table.Append([]string{fmt.Sprintf("%d", dbversion), data2.IndexLocation, data2.BaseDir, data2.Date.Format(time.RFC3339), viper.GetString("dgraph")})
	table.Render()
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

	//	for k := len(data.Entries); k < dbversion; k++ {
	for k := min; k < max; k++ {
		v := L[k]
		table.Append([]string{fmt.Sprintf("%d", k+1), nextstate, v.Filename, "-", v.Description})
	}
	table.Render()
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Migrations info",
	Long:  `Migrations info`,
	Run: func(cmd *cobra.Command, args []string) {
		data, err := internal.GetIndexState()
		if err != nil {
			fmt.Println(err)
			//			return
		} else {
			err = internal.VerifyIndexState(&data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		data2, err := internal.GetDatabaseState()
		if err != nil {
			fmt.Println("Database not initialized!")
		} else {
			infoShowDatabaseMetaInfo(data2)
		}
		fmt.Println()
		if detailed {
			//			infoShowDatabaseLongInfo(data, data2)
			infoShowDatabaseInfo(data, data2)
		} else {
			infoShowDatabaseInfo(data, data2)
		}
	},
}
