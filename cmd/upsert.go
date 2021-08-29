package cmd

import (
	"fmt"
	"strings"

	"github.com/miko/dmt/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	upsertCmd.Flags().IntVarP(&from, "from", "f", 0, "From target")
	rootCmd.AddCommand(upsertCmd)
}

var upsertCmd = &cobra.Command{
	Use:          "upsert",
	Short:        "Execute upsert operation",
	Long:         `Execute upsert operation in a given file without affecting database migrations. Use carefully!`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		vrb := viper.GetBool("verbose")
		var err error
		if vrb {
			fmt.Printf("[info] Started migrator %s\n", VER)
		}

		from := viper.GetString("index")
		if from == "" {
			err = fmt.Errorf("No source file given. Use -i")
			fmt.Println(err)
			return err
		}

		if strings.HasSuffix(from, ".rdf") {
			err = internal.ProcessFile(from, "mutation.rdf")
		} else {
			err = fmt.Errorf("*.rfd file supported, %s given", from)
		}
		return err
	},
}
