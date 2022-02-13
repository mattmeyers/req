package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattmeyers/req"
	"github.com/urfave/cli/v2"
)

type App struct {
	reader *bufio.Reader
	writer io.Writer

	args   []string
	config *req.Config
	app    *cli.App
}

func New(args []string) *App {
	a := &App{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
		args:   args,
	}

	a.app = &cli.App{
		Name: "req",
		Before: func(c *cli.Context) error {
			var err error
			a.config, err = req.ParseConfig("")
			if err != nil {
				return err
			}

			return nil
		},
		Action: a.handleRepl,
		Commands: []*cli.Command{
			{
				Name:   "send",
				Action: a.handleSend,
			},
		},
	}

	return a
}

func (a *App) Run() error {
	return a.app.Run(a.args)
}

func (a *App) handleRepl(c *cli.Context) error {
	fmt.Print("Welcome to the req REPL\n\n")
	for {
		fmt.Print(">> ")

		text, err := a.reader.ReadString('\n')
		if err != nil {
			return err
		}

		text = strings.Trim(text, " \n")
		if text == "" {
			continue
		}

		command := strings.SplitN(text, " ", 2)
		switch command[0] {
		case "send":
			if len(command) != 2 {
				fmt.Println("[Error]: alias of glob required")
				continue
			}

			files, err := a.getFiles(command[1])
			if err != nil {
				fmt.Printf("[Error]: could not retrieve files: %v\n", err)
				continue
			}

			err = a.sendRequests(files)
			if err != nil {
				fmt.Printf("[Error]: could not send request(s): %v\n", err)
				continue
			}

		case "aliases":
			i := 1
			for alias, file := range a.config.Aliases {
				fmt.Printf("%d: %s -> %s\n", i, alias, file)
				i++
			}
		}
	}
}

func (a *App) handleSend(c *cli.Context) error {
	files, err := a.getFiles(c.Args().First())
	if err != nil {
		return fmt.Errorf("[Error]: could not retrieve files: %v", err)
	}

	err = a.sendRequests(files)
	if err != nil {
		return fmt.Errorf("[Error]: could not send request(s): %v", err)
	}

	return nil
}

func (a *App) getFiles(path string) ([]string, error) {
	var files []string
	var err error

	if alias, ok := a.config.Aliases[path]; ok {
		files = append(files, alias)
	} else {
		files, err = filepath.Glob(path)
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

func (a *App) sendRequests(files []string) error {
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

		for _, assertion := range reqfile.Response.Assertions {
			err = assertion.Assert(request, response)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	return nil
}
