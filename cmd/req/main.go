package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattmeyers/req"
	"github.com/urfave/cli/v2"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(argv []string) error {
	app := &cli.App{
		Name: "req",
		Action: func(c *cli.Context) error {
			files, err := filepath.Glob(c.Args().First())
			if err != nil {
				return err
			}

			if len(files) == 0 {
				return errors.New("no reqfiles provided")
			}

			for _, file := range files {
				fmt.Printf("Running %s...\n", file)
				reqfile, err := req.ParseReqfile(file)
				if err != nil {
					return err
				}

				client := req.NewClient()
				request, response, err := client.Do(reqfile.Request)
				if err != nil {
					return err
				}

				for _, assertion := range reqfile.Assertions {
					err = assertion.Assert(request, response)
					if err != nil {
						fmt.Println(err)
					}
				}
			}

			return nil
		},
	}

	return app.Run(argv)
}
