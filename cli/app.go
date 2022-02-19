package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattmeyers/req"
	"github.com/urfave/cli/v2"
)

type App struct {
	reader *bufio.Reader
	writer io.Writer
	logger req.Logger

	args   []string
	config *req.Config
	env    string
	app    *cli.App
}

func New(args []string) *App {
	logger, _ := req.NewLevelLogger(req.LevelDebug, os.Stdout)
	a := &App{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
		logger: logger,
		args:   args,
	}

	a.app = &cli.App{
		Name:  "req",
		Usage: "A CLI/REPL HTTP request runner",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:      "config",
				Aliases:   []string{"c"},
				Usage:     "Point to a reqrc config file",
				Value:     "./.reqrc",
				TakesFile: true,
			},
		},
		Before: func(c *cli.Context) error {
			reqrcPath := c.Path("config")

			var err error
			a.config, err = req.ParseConfig(reqrcPath)
			if err != nil {
				return err
			}

			a.env = a.config.DefaultEnv

			return nil
		},
		Action: a.handleReplCommand,
		Commands: []*cli.Command{
			{
				Name:   "send",
				Usage:  "Send a request by alias or glob",
				Action: a.handleSendCommand,
			},
			{
				Name:   "list",
				Usage:  "List all available requests",
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
	fmt.Fprint(a.writer, "Welcome to the req REPL.\nType help to see available commands.\n\n")
	prompt := ">>"
	if a.env != "" {
		prompt = fmt.Sprintf("[%s] >>", a.env)
	}

	for {
		text, err := a.getInput(prompt)
		if err != nil {
			return err
		} else if text == "" {
			continue
		}

		command := strings.SplitN(text, " ", 2)
		switch command[0] {
		case "send":
			if len(command) != 2 {
				a.logger.Error("alias or glob required")
				continue
			}

			err = a.handleSend(command[1])
			if err != nil {
				a.logger.Error(err.Error())
				continue
			}

		case "list":
			err = a.handleList()
			if err != nil {
				a.logger.Error(err.Error())
			}

		case "new":
			err = a.handleNew()
			if err != nil {
				a.logger.Error(err.Error())
				continue
			}

		case "env":
			for k, v := range a.config.Environments[a.env] {
				fmt.Fprintf(a.writer, "%s = %s\n", k, v)
			}

		case "env-select":
			if len(command) != 2 {
				a.logger.Error("new env required")
				continue
			}

			newEnv := strings.TrimSpace(command[1])
			if _, ok := a.config.Environments[newEnv]; !ok {
				a.logger.Error("env does not exist (create with env-new)")
				continue
			}

			a.env = newEnv

		case "env-new":
			if len(command) != 2 {
				a.logger.Error("new env required")
				continue
			}

			err = a.config.NewEnv(command[1])
			if err != nil {
				a.logger.Error(err.Error())
				continue
			}

			a.env = command[1]

		case "env-set":
			if len(command) != 2 {
				a.logger.Error("key and value required")
				continue
			}

			keyValue := strings.SplitN(command[1], " ", 2)
			if len(keyValue) != 2 {
				a.logger.Error("key and value required")
				continue
			}

			err = a.config.SetEnvValue(a.env, keyValue[0], keyValue[1])
			if err != nil {
				a.logger.Error(err.Error())
				continue
			}

		case "env-delete":
			if len(command) != 2 {
				a.logger.Error("key required")
				continue
			}

			err = a.config.DeleteEnvValue(a.env, command[1])
			if err != nil {
				a.logger.Error(err.Error())
				continue
			}

		case "help", "h":
			a.printHelp()

		case "quit", "q", "exit":
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

func (a *App) handleNew() error {
	method, err := a.getInput("Method:")
	if err != nil {
		return err
	}

	url, err := a.getInput("URL:")
	if err != nil {
		return err
	}

	client := req.NewClient()
	_, res, err := client.Do(req.Request{Method: method, URL: url})
	if err != nil {
		return err
	}

	err = a.printResponse(res)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) getInput(prompt string) (string, error) {
	fmt.Fprintf(a.writer, "%s ", prompt)

	text, err := a.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.Trim(text, " \n"), nil
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

		err = a.printResponse(response)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) printResponse(response *http.Response) error {
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

	return nil
}

func (a *App) printHelp() {
	fmt.Fprint(a.writer, "Available commands:\n")
	fmt.Fprint(a.writer, "  h, help              Display this help message.\n")
	fmt.Fprint(a.writer, "  list                 List all available requests including aliases.\n")
	fmt.Fprint(a.writer, "  send {alias|glob}    Send a request.\n")
	fmt.Fprint(a.writer, "  new    				 Interactively define a new request.\n")
	fmt.Fprint(a.writer, "  env                  Display all values in the current env.\n")
	fmt.Fprint(a.writer, "  env-select {env}     Change the current env.\n")
	fmt.Fprint(a.writer, "  env-new {env}        Create a new env and switch to it.\n")
	fmt.Fprint(a.writer, "  env-set {key} {val}  Set a value in the current env.\n")
	fmt.Fprint(a.writer, "  env-delete {key}     Delete a value from the current env.\n")
	fmt.Fprint(a.writer, "  q, quit, exit        Exit the REPL.\n")
}
