[![build](https://github.com/redpierrot/bestow/actions/workflows/build.yml/badge.svg?branch=main&event=push)](https://github.com/redpierrot/bestow/workflows/build.yml)
[![codecov](https://codecov.io/github/redpierrot/bestow/graph/badge.svg?token=DCC3VHGZTC)](https://codecov.io/github/redpierrot/bestow)
[![last commit](https://shields.io/github/last-commit/redpierrot/bestow)](https://github.com/redpierrot/bestow)
[![Go](https://shields.io/github/go-mod/go-version/redpierrot/bestow)](https://go.dev)

# Bestow

A modern, fast symlink manager. (Be)Stow is the spiritual successor to [GNU stow](https://www.gnu.org/savannah-checkouts/gnu/stow/stow.html) written in [Go](https://go.dev).

Bestow creates and manages symlinks between a source directory and a destination directory.
If you are managing your dotfiles as a repository, bestow is your friend to manage your configs by linking the dotfiles to your config home directory.

## Installation

```sh
go install github.com/redpierrot/bestow@latest
```

Or build from source:

```sh
git clone https://github.com/redpierrot/bestow.git
cd bestow
task build          # requires Task (taskfile.dev)
# or: go build -o bin/bestow .
```

## Quick Start

```sh
# 1. Initialize bestow with your dotfiles directory
bestow init --source ~/dotfiles --destination $HOME

# 2. Stow everything
bestow stow

# 3. Or stow specific packages
bestow stow nvim zsh git
```

## The Package Model

Each top-level directory inside your source is a **package**. Bestow mirrors the directory structure inside each package when creating symlinks — it does not symlink the package directory itself, only the files within it.

```
dotfiles/
├── zsh/
│   └── .zshrc                      → $HOME/.zshrc
├── nvim/
│   └── .config/
│       └── nvim/
│           └── init.lua            → $HOME/.config/nvim/init.lua
└── git/
    └── .gitconfig                  → $HOME/.gitconfig
```

> **NOTE**: Unlike GNU _stow_, _bestow_ does not link directories. Instead, _bestow_ creates the intermediate directories and links only files. This is a design choice to not to overcomplicate the stowing process.

## Commands

### `bestow init`

Initializes bestow by writing a config file and a global ignore file to `$XDG_CONFIG_HOME/bestow/` (defaults to `~/.config/bestow/`).

```sh
bestow init --source ~/dotfiles
bestow init --source ~/dotfiles --destination ~/sandbox   # custom destination
bestow init --source ~/dotfiles --force                   # overwrite existing config
```

| Flag                | Description                                |
| ------------------- | ------------------------------------------ |
| `-s, --source`      | Source directory (required)                |
| `-d, --destination` | Symlink destination (default: `$HOME`)     |
| `--ignore-list`     | Override the global ignore patterns        |
| `-f, --force`       | Overwrite existing config and ignore files |

### `bestow stow`

Creates symlinks from source packages to the destination.

```sh
bestow stow                         # stow all packages
bestow stow nvim git                # stow specific packages nvim and git
bestow stow --force                 # replace conflicting files
bestow stow --backup                # back up conflicting files before replacing
bestow stow --adopt                 # move conflicting files into the source
bestow stow --dry-run               # preview without making changes
```

| Flag                | Description                                                      |
| ------------------- | ---------------------------------------------------------------- |
| `-s, --source`      | Override the source directory                                    |
| `-d, --destination` | Override the destination directory                               |
| `-f, --force`       | Remove existing files and create symlinks                        |
| `-a, --adopt`       | Move existing files into the source, then link                   |
| `-b, --backup`      | Rename existing files to `<file>.N.bestow.backup` before linking |
| `-n, --dry-run`     | Preview operations without touching the filesystem               |

### `bestow unstow`

Removes managed symlinks from the destination.

```sh
bestow unstow                       # unstow all packages
bestow unstow nvim git              # unstow specific packages
bestow unstow --dry-run             # preview
```

| Flag                | Description                                        |
| ------------------- | -------------------------------------------------- |
| `-s, --source`      | Override the source directory                      |
| `-d, --destination` | Override the destination directory                 |
| `-n, --dry-run`     | Preview operations without touching the filesystem |

### Global Flags

| Flag            | Description                                              |
| --------------- | -------------------------------------------------------- |
| `-v, --verbose` | Enable debug-level log output                            |
| `-q, --quiet`   | Suppress all output except errors                        |
| `--profile`     | Use a named profile from the config (default: `default`) |
| `--config-file` | Use a custom config file path                            |

## Configuration

Bestow is configured via `$XDG_CONFIG_HOME/bestow/config.yaml` (default: `~/.config/bestow/config.yaml`).

Running `bestow init -s /home/user/dotfiles` generates this file.

```yaml
version: 0.1.0

profiles:
  default:
    source: /home/user/dotfiles
    destination: /home/user
```

You can define multiple profiles and switch between them with `--profile`:

```yaml
profiles:
  default:
    source: ~/dotfiles
    destination: ~
  sandbox:
    source: ~/dotfiles
    destination: ~/sandbox
```

```sh
bestow stow --profile sandbox
```

Flag values take precedence over the config file. Environment variables use the `BESTOW_` prefix (e.g., `BESTOW_SOURCE`).

## Ignoring Files

Bestow merges ignore patterns from three sources, in order:

| Source              | Location                           | Scope                       |
| ------------------- | ---------------------------------- | --------------------------- |
| Global ignore file  | `~/.config/bestow/.bestowignore`   | All packages                |
| Source ignore file  | `<source>/.bestowignore`           | All packages in this source |
| Package ignore file | `<source>/<package>/.bestowignore` | That package only           |

Patterns follow [doublestar](https://github.com/bmatcuk/doublestar) glob syntax (similar to `.gitignore`). Lines starting with `#` are comments.

```
# ~/.config/bestow/.bestowignore
.git
.gitignore
README.md
LICENSE
**/.bestowignore
**/.stow-local-ignore
```

## Conflict Resolution

When a file already exists at the destination, bestow applies a conflict resolution strategy. The default is **skip**.

| Strategy | Flag        | Behavior                                                  |
| -------- | ----------- | --------------------------------------------------------- |
| Skip     | _(default)_ | Leave the existing file alone                             |
| Force    | `--force`   | Delete the existing file and create the symlink           |
| Backup   | `--backup`  | Rename to `<file>.0.bestow.backup` and create the symlink |
| Adopt    | `--adopt`   | Move the existing file into the source, then link back    |

> **Note:** Bestow never applies a conflict strategy to files it already manages. A destination that already points to the correct source is reported as _up-to-date_ and left untouched.

> **Important**: Bestow welcomes contributions, but do not approve contributions that _seems_ AI-generated garbage. If you don't understand your code, do not submit it to somewhere else.

## License

Author - Thisaru Guruge - All Rights Reversed (ɔ)
