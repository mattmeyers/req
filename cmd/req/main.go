package main

import (
	"fmt"
	"os"

	"github.com/mattmeyers/req/cli"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(argv []string) error {
	return cli.New(argv).Run()
}
