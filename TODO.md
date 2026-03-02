# TODO

## Tier 2 - Enhanced Read Operations
- [ ] `get_server_detail` - Get detailed info for a single server by name or Moid
- [ ] `get_alarm_detail` - Get full alarm details by Moid
- [ ] `list_server_profiles` - List server profiles and templates
- [ ] `list_policies` - List policies by type (NTP, BIOS, boot, etc.)
- [ ] Add `$select` parameter to reduce response payload size
- [ ] Add `$orderby` parameter for sorting results

## Tier 3 - Write Operations
- [ ] `acknowledge_alarm` - Acknowledge an alarm
- [ ] `set_power_state` - Power on/off/cycle a server
- [ ] `assign_profile` - Assign a server profile to a server

## Remote Connections (Streamable HTTP Transport)
- [ ] Add `TRANSPORT` env var to switch between `stdio` (default) and `http` mode
- [ ] Add `PORT` env var (default 8080) for HTTP listen port
- [ ] Use MCP SDK `NewStreamableHTTPHandler` for HTTP transport
- [ ] Add `MCP_API_KEY` env var for optional Bearer token auth on HTTP requests
- [ ] Auth middleware: reject requests without valid `Authorization: Bearer <key>` when MCP_API_KEY is set
- [ ] Add `EXPOSE 8080` to Dockerfile
- [ ] Add `podman` and `podman-run` Makefile targets
- [ ] Document remote connection setup in README (Claude Desktop `mcp-remote`, direct URL, podman run examples)

## Infrastructure
- [ ] Unit tests with mock Intersight responses
- [ ] Integration test with real Intersight (behind build tag)
- [ ] CI/CD pipeline (GitHub Actions: lint, test, build, push image)
- [ ] Release automation (goreleaser)
- [ ] MCP resources for static/cached data (orgs, profiles)
- [ ] MCP prompts for guided workflows (diagnose server, audit firmware)
- [ ] Pagination for large result sets (auto-follow $skip/$top)
- [ ] `INTERSIGHT_API_KEY` env var as alternative to file path (inline key)
