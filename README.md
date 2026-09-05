# Intersight MCP Server

An [MCP](https://modelcontextprotocol.io/) server for [Cisco Intersight](https://intersight.com/), enabling AI assistants to query and manage server infrastructure.

## Features

- List physical servers with hardware details (CPU, memory, model, firmware)
- View active alarms by severity
- Check HCL (Hardware Compatibility List) compliance
- Query running firmware versions
- List organizations
- OData `$filter` and `$orderby` support for server-side queries and sorting
- Truncation-aware results: every listing reports the total match count, so you know when a page is partial

## Quick Start

### Prerequisites

- Cisco Intersight API key ([generate one here](https://intersight.com/an/settings/api-keys/))
- Go 1.25+ (build from source) or Docker/Podman

### Configuration

| Variable | Required | Description |
|---|---|---|
| `INTERSIGHT_API_KEY_ID` | Yes | Your Intersight API key ID |
| `INTERSIGHT_API_KEY_FILE` | Yes | Path to your API private key PEM file |
| `INTERSIGHT_API_HOST` | No | Intersight hostname for Intersight Appliance, with or without scheme (default: `intersight.com`) |

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

All tools accept the same optional parameters:

| Parameter | Description |
|---|---|
| `filter` | OData `$filter` expression. String literals must be single-quoted. |
| `orderby` | OData `$orderby` expression, e.g. `CreationTime desc`. Comma-separate multiple properties. |
| `top` | Maximum results. Defaults to 100 (the Intersight default); values above 1000 are clamped to the API maximum. |

Every tool is annotated `readOnlyHint`, so MCP clients can run them without write-approval prompts.

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

# Newest critical alarms first
filter: "Severity eq 'Critical'", orderby: "CreationTime desc"
```

Intersight supports `$filter`, `$select`, `$top`, `$skip`, `$orderby`, `$count`,
`$inlinecount`, `$expand` and `$apply`. See the
[query syntax reference](https://developer.cisco.com/docs/intersight/query-syntax/).

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
