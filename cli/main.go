package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/asssaf/st7735-go/cli/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Fprintln(flag.CommandLine.Output(), err.Error())
		flag.Usage()
		os.Exit(1)
	}

	os.Exit(0)
}
