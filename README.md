# core-cli

![CI](https://github.com/leaflock/core-cli/actions/workflows/ci.yml/badge.svg)

The `leaf` CLI — the primary command-line interface for the LeafLock platform. Responsible for authentication, environment management, and developer tooling.

---

## Tech Stack

- Go
- [Viper](https://github.com/spf13/viper) — configuration loading
- [golangci-lint](https://golangci-lint.run) — linting

---

## Local Setup

Prerequisites: Go 1.22+, `golangci-lint`.

```bash
git clone https://github.com/leaflock/core-cli.git
cd core-cli
go mod download
make build-dev
```

---

## Environment Variables

| Variable            | Description                                                            | Required | Accepted Values                      |
| ------------------- | ---------------------------------------------------------------------- | -------- | ------------------------------------ |
| `LEAF_ENV`          | Active runtime profile. Overrides the build-time default.              | No       | `dev`, `test`, `prod`                |
| `LEAF_CONFIG_DIR`   | Override the config file search directory. Dev and test profiles only. | No       | Any valid directory path             |
| `LEAF_CONFIG_DEBUG` | Config loader debug output level.                                      | No       | `0` (off), `1` (package), `2` (file) |

---

## Available Commands

Run `make help` to list all available commands.

---

## Deployment and Rollback

See [docs/](docs/) for deployment and rollback specifics for this repo.
