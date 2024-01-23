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
	VER       = "v0.2.30"
)

var (
	// Used for flags.
	dgraph  string
	index   string // index file/dir - json?
	verbose bool

	rootCmd = &cobra.Command{
		Use:   "dmt",
		Short: "DGraph migration tool",
		Long:  `DGraph migration tool ` + VER,
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

	rootCmd.PersistentFlags().StringP("index", "i", internal.OrDefault(os.Getenv(envPrefix+"_INDEX"), ""), "Migration index file")
	viper.BindPFlag("index", rootCmd.PersistentFlags().Lookup("index"))
	viper.SetDefault("index", internal.OrDefault(os.Getenv(envPrefix+"_INDEX"), ""))

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose mode")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.SetDefault("verbose", false)

	rootCmd.PersistentFlags().StringP("export_url", "u", internal.OrDefault(os.Getenv(envPrefix+"_EXPORT_URL"), ""), "Export URL (S3)")
	viper.BindPFlag("export_url", rootCmd.PersistentFlags().Lookup("export_url"))
	viper.SetDefault("export_url", internal.OrDefault(os.Getenv(envPrefix+"_EXPORT_URL"), ""))

	rootCmd.PersistentFlags().StringP("export_access_key", "a", internal.OrDefault(os.Getenv(envPrefix+"_EXPORT_ACCESS_KEY"), ""), "Export AccessKey (S3)")
	viper.BindPFlag("export_access_key", rootCmd.PersistentFlags().Lookup("export_access_key"))
	viper.SetDefault("export_access_key", internal.OrDefault(os.Getenv(envPrefix+"_EXPORT_ACCESS_KEY"), ""))

	rootCmd.PersistentFlags().StringP("export_secret_key", "k", internal.OrDefault(os.Getenv(envPrefix+"_EXPORT_SECRET_KEY"), ""), "Export SecretKey (S3)")
	viper.BindPFlag("export_secret_key", rootCmd.PersistentFlags().Lookup("export_secret_key"))
	viper.SetDefault("export_secret_key", internal.OrDefault(os.Getenv(envPrefix+"_EXPORT_SECRET_KEY"), ""))

	rootCmd.PersistentFlags().StringP("header", "H", internal.OrDefault(os.Getenv(envPrefix+"_HEADER"), ""), "Additional HTTP header for graphql data - in form of k=v")
	viper.BindPFlag("header", rootCmd.PersistentFlags().Lookup("header"))
	viper.SetDefault("header", internal.OrDefault(os.Getenv(envPrefix+"_HEADER"), ""))

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
