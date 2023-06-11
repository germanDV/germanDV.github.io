---
title: a-makefile-for-go-projects
published: 2022-12-02
revision: 2022-12-02
excerpt: I like using a Makefile in my Go projects. These are some tasks that I find useful in pretty much all of them.
---

Let's start with a simple task, create a _Makefile_ at the root of your project with the contents:

```makefile
dev:
  air .
```

You can now run `make dev` and it will execute the command `air .`.

[Air](https://github.com/cosmtrek/air) is a tool I use for hot-reloading during development, but of course you can run any other command you wish.

Makefiles are supposed to deal with files, in our case, we are using it as a task runner, so, we will add the `.PHONY` target to be explicit about that fact:

```makefile
.PHONY: dev
dev:
  @echo "Starting web server in development mode"
  ENV=development air .
```

The `@` at the beginning of the _echo_ command means "do not print this line", which would be redundant for _echo_.

I have added an environment variable `ENV`, it's important that you do this in the same line as the command that the variable is supposed to affect (`air` in this case). If you set the environment variable in one line and run the command in another one, it won't work.

This won't work as expected:

```makefile
.PHONY: dev
dev:
  @echo "Starting web server in development mode"
  export ENV=development
  air .
```

If you run `make` with no arguments, it will run the first task in the Makefile. So, I think it's a good idea to put a _help_ task as the first one, which will print a small description of every task available:

```makefile
## help: print this help message.
.PHONY: help
help:
  @echo 'Usage:'
  @sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
```

For any task that you wish to show in the help message, you just need to write a comment with the format `## <task>: <message>`.

Let's add a help message to our _dev_ task:

```makefile
## dev: run with hot-reloading.
.PHONY: dev
dev:
  @echo "Starting web server in development mode"
  export ENV=development
  air .
```

If you now run `make` or `make help`, you will get:

```
Usage:
  help print this help message.
  dev  run with hot-reloading.
```

Usually, in my Go projects I would use tools that are not part of the application itself, they are not listed in _go.mod_ so to say. For example, here we are using _air_ for hot-reloading, and if I were to work with a database, I would use something like [migrate](https://github.com/golang-migrate/migrate) to handle database migrations.

I like having one task that consolidates these dependencies, so that it's clear for developers joining the project what tools are used, and they can install them with a single command:

```makefile
## deps: install external dependencies not used in source code
.PHONY: deps
deps:
  @echo 'Installing `air` for hot-reloading'
  go install github.com/cosmtrek/air@latest
```

Since this is going to install things on users' systems, I think it would be nice to ask for confirmation before proceeding. We can have a little _confirm_ helper that we can attach to any task we wish:

```makefile
.PHONY: confirm
confirm:
  @echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]
```

Notice the lack of help message, that's because this is a task for internal use only.

To ask for confirmation, we simply need to add the _confirm_ task. Let's add it to _deps_:

```makefile
## deps: install external dependencies not used in source code
.PHONY: deps
deps: confirm
  @echo 'Installing `air` for hot-reloading'
  go install github.com/cosmtrek/air@latest
```

If you run `make deps`, you will get:

```
Are you sure? [y/N]
```

This wouldn't be complete without adding a task for tests. In addition to tests, we will also tidy and verify dependencies, fmt and vet the code:

```makefile
## audit: tidy dependencies, format, vet and test.
.PHONY: audit
audit:
  @echo 'Tidying and verifying module dependencies...'
  go mod tidy
  go mod verify
  @echo 'Formatting code...'
  go fmt ./...
  @echo 'Vetting code...'
  go vet ./...
  @echo 'Running tests...'
  ENV=testing go test -race -vet=off ./...
```

Another thing that might come in handy in a makefile, is to read an environemt variable and use a default if it is not present.
We can achieve that with the following syntax:

```makefile
SOME_VAR ?= 'i_am_a_default'
```

We can then use it in any task:

```makefile
SOME_VAR ?= 'i_am_a_default'

.PHONY: example
example:
  @echo 'Value of SOME_VAR is: ${SOME_VAR}'
```

If you run `make example`, you will get `Value of SOME_VAR is: i_am_a_default`.
If you run `SOME_VAR=injected make example`, you will get `Value of SOME_VAR is: injected`.

I would normally have one or more build tasks, maybe one to build for the current arch and another one to build for all targets. If I'm working with a database, I would also have tasks to deal with creating, applying and reverting migrations. But these are more project-specific so I won't include them here.

To sum up, this is the entire Makefile:

```makefile
## help: print this help message.
.PHONY: help
help:
  @echo 'usage:'
  @sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
  @echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

## audit: tidy dependencies, format, vet and test.
.PHONY: audit
audit:
  @echo 'Tidying and verifying module dependencies...'
  go mod tidy
  go mod verify
  @echo 'Formatting code...'
  go fmt ./...
  @echo 'Vetting code...'
  go vet ./...
  @echo 'Running tests...'
  ENV=testing go test -race -vet=off ./...

## dev: run with hot-reloading.
.PHONY: dev
dev:
  ENV=development air .

## deps: install external dependencies not used in source code
.PHONY: deps
deps: confirm
  @echo 'Installing `air` for hot-reloading'
  go install github.com/cosmtrek/air@latest
```

Some of this I have stolen from two great books by Alex Edwards:

- [Let's Go](https://lets-go.alexedwards.net/)
- [Let's Go Further](https://lets-go-further.alexedwards.net/)
