package main

import (
	"fmt"
	"os"
	"runtime/debug"

	_ "github.com/joho/godotenv/autoload"
	"github.com/miko/dmt/cmd"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			fmt.Printf("Error was: %s\n", r)
			os.Exit(1)
		}
	}()

	cmd.Execute()
}
