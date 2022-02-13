package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattmeyers/req"
	"github.com/mattmeyers/req/log"
	"github.com/urfave/cli/v2"
)

type App struct {
	reader *bufio.Reader
	writer io.Writer
	logger log.Logger

	args   []string
	config *req.Config
	app    *cli.App
}

func New(args []string) *App {
	logger, _ := log.NewLevelLogger(log.LevelDebug, os.Stdout)
	a := &App{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
		logger: logger,
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
	fmt.Fprint(a.writer, "Welcome to the req REPL\n\n")
	for {
		fmt.Fprint(a.writer, ">> ")

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
		case ":send":
			if len(command) != 2 {
				a.logger.Error("alias of glob required")
				continue
			}

			files, err := a.getFiles(command[1])
			if err != nil {
				a.logger.Error("could not retrieve files: %v\n", err)
				continue
			} else if len(files) == 0 {
				a.logger.Error("no files specified\n")
				continue
			}

			err = a.sendRequests(files)
			if err != nil {
				a.logger.Error("could not send request(s): %v\n", err)
				continue
			}

		case ":list":
			glob := fmt.Sprintf("%s/*.hcl", strings.TrimRight(a.config.Root, "/"))
			files, err := filepath.Glob(glob)
			if err != nil {
				return err
			}

			aliasLookup := make(map[string]string)
			for k, v := range a.config.Aliases {
				aliasLookup[v] = k
			}

			for i, file := range files {
				if alias, ok := aliasLookup[file]; ok {
					fmt.Fprintf(a.writer, "%d: %s -> %s\n", i+1, alias, file)
				} else {
					fmt.Fprintf(a.writer, "%d: %s\n", i+1, file)
				}
			}
		}
	}
}

func (a *App) handleSend(c *cli.Context) error {
	files, err := a.getFiles(c.Args().First())
	if err != nil {
		a.logger.Error("could not retrieve files: %v", err)
		return nil
	}

	err = a.sendRequests(files)
	if err != nil {
		a.logger.Error("could not send request(s): %v", err)
		return nil
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
		a.logger.Info("Running %s...\n", file)
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
