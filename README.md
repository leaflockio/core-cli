# core-cli

![CI](https://github.com/leaflockio/core-cli/actions/workflows/ci.yml/badge.svg)

The `leaf` CLI — developer experience tooling by LeafLock.

---

## Tech Stack

- [Go](https://go.dev)
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Viper](https://github.com/spf13/viper) — configuration loading
- [golangci-lint](https://golangci-lint.run) — linting

---

## Local Setup

Prerequisites: Go 1.26+, `golangci-lint`.

```bash
git clone https://github.com/leaflockio/core-cli.git
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

List all available commands using:

```bash
make help
```

---

## Documentation

See [docs/](docs/).
