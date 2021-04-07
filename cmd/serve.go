package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	port int
)

func init() {
	serveCmd.Flags().IntVar(&port, "port", 8080, "Listen on port")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serving as web service",
	Long:  `All software has serves. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Printf("Listening at port %d \n", port)
		fmt.Println("Not implemented")
	},
}
