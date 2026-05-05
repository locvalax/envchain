# envchain

Manage and inject environment variables from multiple sources with precedence rules.

## Installation

```bash
go install github.com/yourusername/envchain@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/envchain.git && cd envchain && go build ./...
```

## Usage

Define your sources in an `envchain.yaml` file and run any command with the resolved environment:

```yaml
sources:
  - type: dotenv
    path: .env.local
  - type: vault
    path: secret/myapp
  - type: env
```

```bash
# Inject resolved variables into a command
envchain run -- ./myapp

# Print resolved environment without running a command
envchain dump

# Specify a custom config file
envchain --config ./config/envchain.yaml run -- npm start
```

Sources are merged in order, with **later sources taking higher precedence**. Variables already set in the environment can be configured to override or defer to upstream sources.

## Configuration

| Field      | Description                          |
|------------|--------------------------------------|
| `type`     | Source type (`dotenv`, `env`, `vault`, `ssm`) |
| `path`     | Path or identifier for the source    |
| `optional` | Skip source if unavailable           |

## License

MIT — see [LICENSE](LICENSE) for details.