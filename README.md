# req

`req` is a command line HTTP request runner that uses HCL for request definitions and provides a REPL for easy usage.

## Installation

`req` can be directly installed with the following command

```sh
go install github.com/mattmeyers/req/cmd/req@latest
```

## Configuration

When the `req` command is invoked, it looks for a file named `.reqrc` in the current working directory. If found, it is loaded into memory and used to configure the session. If not found, `req` falls back to a default configuration. The `.reqrc` file is a TOML file that allows the following values.

```toml
# Sets the root directory that req uses to look for reqfiles in.
root = ''

# Sets the default environment to use in the REPL. This env must be defined
# below.
default_env = ''

# A table of request aliases. These values can be used to quickly refer
# to a specific request. Aliases must be defined with a full path relative
# to the directory containing the .reqrc file. The root configuration value
# is not applied to alias paths.
[aliases]

# A list of environments. An environment is a table of key/value pairs that
# can be accessed in request templates. For simplicity, all values MUST be
# strings.
[environments.<env_name>]
```

A sample `.reqrc` is as follows.

```toml
root = './requests/'
default_env = 'local'

[aliases]
echo = './requests/echo.hcl'
ping = './requests/ping.hcl'

[environments.local]
base_url = 'http://localhost:8080'

[environments.prod]
base_url = 'http://localhost:9001'
```

## Reqfiles

A reqfile is an HCL file that contains a request definition. This file can be loaded by `req` to build and send a request object. The file takes the following schema.

```hcl
request {
    # The HTTP request method.
    method = ""

    # The full URL to make the request to.
    url = ""

    # A map of key/value pairs that define the request headers. Note that
    # all values are expected to be strings.
    headers = {}

    # The request body. This value is a string and is delivered as is without
    # any manipulation. To minimize size, this can be a minified string. To
    # maximize readability, this can take advantage of heredoc syntax.
    body = <<-BODY
    BODY
}
```

To make these request definitions dynamic, HIL interpolation can be used to inject values. At this time, the following variables are injected into the template's context.

- `env`: The current environment's values.

A sample reqfile follows.

```hcl
request {
    method = "POST"
    url = "${env.base_url}/echo"
    headers = {
        Content-Type = "application/json"
    }
    body = <<-BODY
        {
            "foo": "bar"
        }
    BODY
}
```

## Usage

The `req` tool can be used in CLI or REPL mode. Core functionality is available in either mode, but the overall usage differs slightly. CLI mode is intended for single, quick requests. As such, the majority of its configuration is delegated to editing the reqfiles. REPL mode, on the other hand, is intended for longer, multi-request sessions. As such, this mode provides additional commands to manipulate the configuration.

### CLI Usage

```
NAME:
   req - A CLI/REPL HTTP request runner

USAGE:
   req [global options] command [command options] [arguments...]

COMMANDS:
   send     Send a request by alias or glob
   list     List all available requests
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value  Point to a reqrc config file (default: "./.reqrc")
   --help, -h                show help (default: false)
```

### REPL Usage

The REPL prompt takes the form

```
[env]>> command [args...]
```

The prompt shows the current environment name on the left if an environment is selected. Note that an environment is not required to use the REPL. The available commands can be retrieved by typing `help` or `h` and hitting enter. The current list of commands is

```
Available commands:
  h, help              Display this help message.
  list                 List all available requests including aliases.
  send {alias|glob}    Send a request.
  new                                    Interactively define a new request.
  env                  Display all values in the current env.
  env-select {env}     Change the current env.
  env-new {env}        Create a new env and switch to it.
  env-set {key} {val}  Set a value in the current env.
  env-delete {key}     Delete a value from the current env.
  q, quit, exit        Exit the REPL.
```


## Examples

This repository ships with a complete example including a server, `.reqrc`, and reqfiles. To spin up the server, ensure the `go` binary is in your `PATH`, then run

```sh
$ cd examples
$ go run main.go
```

This will spin up a basic server on `127.0.0.1:8080` with two endpoints:

- `GET /ping`
- `POST /echo`

Assuming `req` has been installed and is available in the `PATH`, The CLI mode can be used to run commands such as

```sh
$ req list
$ req send echo
$ req send requests/*
```

REPL mode can also be entered with

```sh
$ req
```