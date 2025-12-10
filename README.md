# Git Opener

A simple terminal-based application to manage and open your Git projects in tmux sessions.

## About

Git Opener is a Go application that scans your Git repositories directory and allows you to quickly:
- Navigate through your projects
- Open a selected project in a tmux session
- Automatically set up vim and terminal windows

The application creates a consistent development environment for each project with:
- Window 1: Vim editor for coding
- Window 2: Terminal for running commands

## Prerequisites

- Go 1.16 or later
- tmux
- vim

## Building the Application

To build the binary locally, make sure Go (1.16 or newer) is installed, then run:

```bash
go build -o git-opener ./src
```

The compiled `git-opener` executable will be created in the repository root and can be copied anywhere on your PATH.

Alternatively, use the provided Makefile for a repeatable build/test workflow:

```bash
make build   # builds git-opener
make run     # builds and starts git-opener
make test    # runs go test ./...
```

## Downloading from Releases

Each tagged release attaches a prebuilt `git-opener` binary. Download the latest version with curl:

```bash
curl -L https://github.com/RollerSweet/git-opener/releases/latest/download/git-opener -o git-opener
chmod +x git-opener
```

You can then run the downloaded binary directly or move it to a location on your PATH (for example `/usr/local/bin`).

## Configuration

The application looks for your Git repositories in the following order:

1. Path specified in the `GIT_REPOS_PATH` environment variable
2. Default path at `~/git` if the environment variable is not set

To specify a custom Git repositories directory:

```bash
export GIT_REPOS_PATH=/path/to/your/git/repos
```

## Running the Application

After building, run the application:

```bash
./git-opener
```

Or with a custom Git repositories path:

```bash
export GIT_REPOS_PATH=/path/to/your/git/repos
./git-opener
```

## Usage

- Use **arrow keys** or **j/k** to navigate through your projects
- Press **Enter** to open the selected project in tmux
- Press **?** for help
- Press **Esc** to exit the application

## Credits

Big thanks to shZsektor (Sharef Zubidat) for the first version and idea - this project just adds a few extra features to that tool.

## Logs

Application logs are stored in `/tmp/app.log` for troubleshooting purposes.

## Automated Releases

A GitHub Actions workflow (`.github/workflows/release.yml`) builds and tests the project on every push to `main`.  
When the latest commit message follows the prefixes below, the pipeline automatically bumps the semantic version, pushes a tag, and publishes a GitHub Release that attaches the compiled `git-opener` binary:

- `fix:` → patch version bump
- `feat:` → minor version bump
- `!feat:` → major version bump (breaking change)

Any push whose HEAD commit does not use one of the prefixes above will run the tests/build but will not cut a release.
