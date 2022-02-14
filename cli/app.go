package cli

import (
	"bufio"
	"bytes"
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
	env    string
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
		Action: a.handleReplCommand,
		Commands: []*cli.Command{
			{
				Name:   "send",
				Action: a.handleSendCommand,
			},
			{
				Name:   "list",
				Action: a.handleListCommand,
			},
		},
	}

	return a
}

func (a *App) Run() error {
	return a.app.Run(a.args)
}

func (a *App) handleReplCommand(c *cli.Context) error {
	fmt.Fprint(a.writer, "Welcome to the req REPL\n\n")
	for {
		if a.env != "" {
			fmt.Fprintf(a.writer, "[%s] >> ", a.env)
		} else {
			fmt.Fprint(a.writer, ">> ")
		}

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
				a.logger.Error("alias or glob required")
				continue
			}

			err = a.handleSend(command[1])
			if err != nil {
				a.logger.Error(err.Error())
				continue
			}

		case ":list":
			err = a.handleList()
			if err != nil {
				a.logger.Error(err.Error())
			}

		case ":env":
			if len(command) != 2 {
				a.logger.Error("new env required")
				continue
			}

			a.env = strings.TrimSpace(command[1])

		case ":quit", ":q", ":exit":
			return nil
		}
	}
}

func (a *App) handleSendCommand(c *cli.Context) error {
	if c.Args().Len() == 0 {
		a.logger.Error("alias or glob required")
		return nil
	}

	err := a.handleSend(c.Args().First())
	if err != nil {
		a.logger.Error(err.Error())
	}

	return nil
}

func (a *App) handleListCommand(c *cli.Context) error {
	err := a.handleList()
	if err != nil {
		a.logger.Error(err.Error())
	}

	return nil
}

func (a *App) handleSend(glob string) error {
	files, err := a.getFiles(glob)
	if err != nil {
		return fmt.Errorf("could not retrieve files: %v", err)
	}

	err = a.sendRequests(files)
	if err != nil {
		return fmt.Errorf("could not send request(s): %v", err)
	}

	return nil
}

func (a *App) handleList() error {
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
		reqfile, err := req.ParseReqfile(file, a.config.Environments[a.env])
		if err != nil {
			return err
		}

		client := req.NewClient()
		_, response, err := client.Do(reqfile.Request)
		if err != nil {
			return err
		}

		a.logger.Info("Got response...\n\n")
		fmt.Fprintf(a.writer, "\t%s %s\n", response.Proto, response.Status)
		for k := range response.Header {
			for _, v := range response.Header.Values(k) {
				fmt.Fprintf(a.writer, "\t%s: %s\n", k, v)
			}
		}
		fmt.Fprint(a.writer, "\n")

		buf, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		if len(buf) > 0 {
			buf = bytes.ReplaceAll(buf, []byte{'\n'}, []byte{'\n', '\t'})
			fmt.Fprintf(a.writer, "\t%s\n", buf)
		}
	}

	return nil
}
