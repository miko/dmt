package main

import (
	"fmt"
	"runtime/debug"

	_ "github.com/joho/godotenv/autoload"
	"github.com/miko/dmt/cmd"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			fmt.Printf("Error was: %s\n", r)
		}
	}()

	cmd.Execute()
}
