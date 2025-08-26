package main

import (
	"context"
	"fmt"
	"mault/cmd"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()
	
	if err := cmd.Mault.RunContext(ctx, os.Args); err != nil {
		fmt.Printf("ERROR: %v\n", err.Error())
	}
}
