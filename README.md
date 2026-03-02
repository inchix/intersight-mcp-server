# Intersight MCP Server

An [MCP](https://modelcontextprotocol.io/) server for [Cisco Intersight](https://intersight.com/), enabling AI assistants to query and manage server infrastructure.

## Features

- List physical servers with hardware details (CPU, memory, model, firmware)
- View active alarms by severity
- Check HCL (Hardware Compatibility List) compliance
- Query running firmware versions
- List organizations
- OData `$filter` support for powerful server-side queries

## Quick Start

### Prerequisites

- Cisco Intersight API key ([generate one here](https://intersight.com/an/settings/api-keys/))
- Go 1.24+ (build from source) or Docker/Podman

### Configuration

| Variable | Required | Description |
|---|---|---|
| `INTERSIGHT_API_KEY_ID` | Yes | Your Intersight API key ID |
| `INTERSIGHT_API_KEY_FILE` | Yes | Path to your API private key PEM file |
| `INTERSIGHT_API_HOST` | No | Intersight hostname (default: `intersight.com`) |

### Build and Run

```bash
make build
export INTERSIGHT_API_KEY_ID="your-key-id"
export INTERSIGHT_API_KEY_FILE="/path/to/key.pem"
./intersight-mcp-server
```

### Docker / Podman

```bash
make docker
docker run --rm \
  -e INTERSIGHT_API_KEY_ID="your-key-id" \
  -e INTERSIGHT_API_KEY_FILE="/app/key.pem" \
  -v /path/to/key.pem:/app/key.pem:ro \
  ghcr.io/inchix/intersight-mcp-server:latest
```

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "intersight": {
      "command": "/path/to/intersight-mcp-server",
      "env": {
        "INTERSIGHT_API_KEY_ID": "your-key-id",
        "INTERSIGHT_API_KEY_FILE": "/path/to/key.pem"
      }
    }
  }
}
```

### Claude Code

Add to your MCP settings:

```json
{
  "mcpServers": {
    "intersight": {
      "command": "/path/to/intersight-mcp-server",
      "env": {
        "INTERSIGHT_API_KEY_ID": "your-key-id",
        "INTERSIGHT_API_KEY_FILE": "/path/to/key.pem"
      }
    }
  }
}
```

## Available Tools

| Tool | Description |
|------|-------------|
| `list_servers` | List physical server inventory with hardware details |
| `list_alarms` | List active alarms with severity and affected objects |
| `list_hcl_statuses` | Check HCL compliance across servers |
| `list_firmware` | List running firmware versions |
| `list_organizations` | List organizations in the account |

All tools except `list_organizations` accept optional `filter` (OData `$filter`) and `top` (max results) parameters.

## OData Filter Examples

```
# Servers by model
filter: "Model eq 'UCSC-C220-M5SX'"

# Powered-on servers only
filter: "OperPowerState eq 'on'"

# Critical alarms
filter: "Severity eq 'Critical'"

# Non-compliant HCL
filter: "Status ne 'Validated'"

# Servers with specific tag
filter: "Tags/any(t:t/Key eq 'Site' and t/Value eq 'London')"
```

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run linter (requires golangci-lint)
make docker   # Build container image
make clean    # Remove binary
```

## License

Apache-2.0. See [LICENSE](LICENSE).
