package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattmeyers/repl"
	"github.com/mattmeyers/reql"
	"github.com/urfave/cli/v2"
)

type App struct {
	reader *bufio.Reader
	writer io.Writer
	logger reql.Logger

	args   []string
	config *reql.Config
	env    string
	app    *cli.App
}

func New(args []string) *App {
	a := &App{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
		args:   args,
	}

	a.app = &cli.App{
		Name:  "reql",
		Usage: "A CLI/REPL HTTP request runner",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:      "config",
				Aliases:   []string{"c"},
				Usage:     "Point to a reqlrc config file",
				Value:     "./.reqlrc",
				TakesFile: true,
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Log additional information",
			},
			&cli.BoolFlag{
				Name:    "very-verbose",
				Aliases: []string{"vv"},
				Usage:   "Log debug information information",
			},
		},
		Before: func(c *cli.Context) error {
			var err error

			reqlrcPath := c.Path("config")
			a.config, err = reql.ParseConfig(reqlrcPath)
			if err != nil {
				return err
			}

			a.env = a.config.DefaultEnv

			level := reql.LevelWarn
			if c.Bool("verbose") {
				level = reql.LevelInfo
			} else if c.Bool("very-verbose") {
				level = reql.LevelDebug
			}

			a.logger, err = reql.NewLevelLogger(level, os.Stdout)
			if err != nil {
				return err
			}

			return nil
		},
		Action: a.handleReplCommand,
		Commands: []*cli.Command{
			{
				Name:  "send",
				Usage: "Send a request by alias or glob",
				Before: func(c *cli.Context) error {
					if env := c.String("env"); env != "" {
						if _, ok := a.config.Environments[env]; !ok {
							return errors.New("unknown env")
						}

						a.env = env
					}

					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "env",
						Aliases: []string{"e"},
						Usage:   "Select the env to use",
					},
				},
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
	r := repl.New().
		WithPrompt(a.prompt).
		WithPreRunHook(func(c *repl.Context) (string, error) {
			return "Welcome to the reql REPL.\nType help to see available commands.\n\n", nil
		}).
		WithHandler(func(c *repl.Context) (string, error) {
			if !strings.HasPrefix(c.Input, "send") {
				return "", repl.ErrNoMatch
			}

			command := strings.SplitN(c.Input, " ", 2)
			if len(command) != 2 {
				return "", repl.NewError("alias or glob required")
			}

			err := a.handleSend(command[1])
			if err != nil {
				return "", repl.NewError(err.Error())
			}

			return "", nil
		}).
		WithHandler(func(c *repl.Context) (string, error) {
			if c.Input != "list" {
				return "", repl.ErrNoMatch
			}

			err := a.handleList()
			if err != nil {
				return "", repl.NewError(err.Error())
			}

			return "", nil
		}).
		WithHandler(func(c *repl.Context) (string, error) {
			if c.Input != "new" {
				return "", repl.ErrNoMatch
			}

			err := a.handleNew()
			if err != nil {
				return "", repl.NewError(err.Error())
			}

			return "", nil
		}).
		WithHandler(func(c *repl.Context) (string, error) {
			if c.Input != "env" {
				return "", repl.ErrNoMatch
			}

			for k, v := range a.config.Environments[a.env] {
				fmt.Fprintf(a.writer, "%s = %s\n", k, v)
			}

			return "", nil
		}).
		WithHandler(func(c *repl.Context) (string, error) {
			if c.Input != "help" && c.Input != "h" {
				return "", repl.ErrNoMatch
			}

			a.printHelp()

			return "", nil
		}).
		WithHandler(func(c *repl.Context) (string, error) {
			if !strings.HasPrefix(c.Input, "env-select") {
				return "", repl.ErrNoMatch
			}

			command := strings.SplitN(c.Input, " ", 2)
			if len(command) != 2 {
				return "", repl.NewError("new env required")
			}

			newEnv := strings.TrimSpace(command[1])
			if _, ok := a.config.Environments[newEnv]; !ok {
				return "", repl.NewError("env does not exist (create with env-new)")
			}

			a.env = newEnv

			return "", nil
		}).
		WithHandler(func(c *repl.Context) (string, error) {
			if !strings.HasPrefix(c.Input, "env-new") {
				return "", repl.ErrNoMatch
			}

			command := strings.SplitN(c.Input, " ", 2)
			if len(command) != 2 {
				return "", repl.NewError("new env required")
			}

			err := a.config.NewEnv(command[1])
			if err != nil {
				return "", repl.NewError(err.Error())
			}

			a.env = command[1]

			return "", nil
		}).
		WithHandler(func(c *repl.Context) (string, error) {
			if !strings.HasPrefix(c.Input, "env-set") {
				return "", repl.ErrNoMatch
			}

			command := strings.SplitN(c.Input, " ", 2)
			if len(command) != 2 {
				return "", repl.NewError("key and value required")
			}

			keyValue := strings.SplitN(command[1], " ", 2)
			if len(keyValue) != 2 {
				return "", repl.NewError("key and value required")
			}

			err := a.config.SetEnvValue(a.env, keyValue[0], keyValue[1])
			if err != nil {
				return "", repl.NewError(err.Error())
			}

			return "", nil
		}).
		WithHandler(func(c *repl.Context) (string, error) {
			if !strings.HasPrefix(c.Input, "env-delete") {
				return "", repl.ErrNoMatch
			}

			command := strings.SplitN(c.Input, " ", 2)
			if len(command) != 2 {
				return "", repl.NewError("key required")
			}

			err := a.config.DeleteEnvValue(a.env, command[1])
			if err != nil {
				return "", repl.NewError(err.Error())
			}

			return "", nil
		}).
		WithHandler(func(c *repl.Context) (string, error) {
			if c.Input != "quit" && c.Input != "q" && c.Input != "exit" {
				return "", repl.ErrNoMatch
			}

			return "", repl.ErrExit
		})

	return r.Run()
}

func (a *App) prompt(ctx *repl.Context) (string, error) {
	prompt := ">> "
	if a.env != "" {
		prompt = fmt.Sprintf("[%s] >> ", a.env)
	}

	return prompt, nil
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

	client := reql.NewClient()
	_, res, err := client.Do(reql.Request{Method: method, URL: url})
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
		reqfile, err := reql.ParseReqfile(file, a.config.Environments[a.env])
		if err != nil {
			return err
		}

		client := reql.NewClient()
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
	fmt.Fprintf(a.writer, "%s %s\n", response.Proto, response.Status)
	for k := range response.Header {
		for _, v := range response.Header.Values(k) {
			fmt.Fprintf(a.writer, "%s: %s\n", k, v)
		}
	}
	fmt.Fprint(a.writer, "\n")

	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if len(buf) > 0 {
		fmt.Fprintf(a.writer, "%s\n", buf)
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
