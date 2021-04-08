package cmd

import (
	"errors"
	"fmt"

	"github.com/miko/dmt/internal"
	"github.com/spf13/cobra"
	//	"github.com/spf13/viper"
)

func init() {
	//	buildCmd.Flags().BoolVarP(&color, "color", "c", false, "Color mode")
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:          "build directoryname",
	Short:        "Build DMT index for given local direcory",
	SilenceUsage: true,
	Long:         `Build DMT index for given local direcory`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a directory name")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		err := internal.BuildIndex(args[0])

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Index rebuild")
	},
}
