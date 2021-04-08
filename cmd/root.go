package cmd

import (
	"os"

	"github.com/miko/dmt/internal"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	envPrefix = "DMT"
)

var (
	// Used for flags.
	dgraph  string
	index   string // index file/dir - json?
	verbose bool

	rootCmd = &cobra.Command{
		Use:   "dmt",
		Short: "DGraph migration tool",
		Long:  `DGraph migration tool v0.1.10.`,
	}
)

// Execute executes the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)
	/*
		rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", internal.OrDefault(os.Getenv(envPrefix+"_CONFIG"), ""), "config file")
		viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
		viper.SetDefault("config", internal.OrDefault(os.Getenv(envPrefix+"_CONFIG"), ""))
	*/
	rootCmd.PersistentFlags().StringP("dgraph", "d", internal.OrDefault(os.Getenv(envPrefix+"_DGRAPH"), "dgraph:9080"), "dgraph server and port")
	viper.BindPFlag("dgraph", rootCmd.PersistentFlags().Lookup("dgraph"))
	viper.SetDefault("dgraph", internal.OrDefault(os.Getenv(envPrefix+"_DGRAPH"), "dgraph:9080"))

	rootCmd.PersistentFlags().StringP("graphql", "g", internal.OrDefault(os.Getenv(envPrefix+"_GRAPHQL"), "http://dgraph:8080"), "dgraph server and port")
	viper.BindPFlag("graphql", rootCmd.PersistentFlags().Lookup("graphql"))
	viper.SetDefault("graphql", internal.OrDefault(os.Getenv(envPrefix+"_GRAPHQL"), "http://dgraph:8080"))

	rootCmd.PersistentFlags().StringP("index", "i", internal.OrDefault(os.Getenv(envPrefix+"_INDEX"), "dmt.json"), "Migration index file")
	viper.BindPFlag("index", rootCmd.PersistentFlags().Lookup("index"))
	viper.SetDefault("index", internal.OrDefault(os.Getenv(envPrefix+"_INDEX"), "dmt.json"))

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose mode")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.SetDefault("verbose", false)

}

func initConfig() {
	cfgFile := internal.OrDefault(os.Getenv(envPrefix+"_CONFIG"), ".dmt.yml")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		//fmt.Printf("Setting config file %s\n", cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".dmt")
	}

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	/*
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		} else {
			fmt.Println("Not using config file")
		}
	*/
}
