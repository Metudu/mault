package main

import (
	"fmt"
	"mault/cmd"
	"os"
)

func main() {
	// Run the application
	if err := cmd.Mault.Run(os.Args); err != nil {
		fmt.Printf("ERROR: %v\n", err.Error())
	}
}
