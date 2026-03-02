package intersight

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FilterArgs is the shared input schema for tools that support OData filtering.
type FilterArgs struct {
	Filter string `json:"filter,omitempty" jsonschema:"OData $filter expression, e.g. Model eq UCSC-C220-M5SX or Severity eq Critical"`
	Top    int32  `json:"top,omitzero" jsonschema:"Maximum number of results to return, default 100, max 1000"`
}

// NoArgs is the input schema for tools that take no parameters.
type NoArgs struct{}

// RegisterTools adds all Intersight MCP tools to the server.
func RegisterTools(s *mcp.Server, c *Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_servers",
		Description: "List physical server inventory from Intersight. Returns name, model, serial, power state, CPU/memory, management IP, and firmware. Supports OData $filter.",
	}, c.listServers)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_alarms",
		Description: "List active alarms from Intersight. Returns severity, description, affected object, and creation time. Supports OData $filter (e.g. \"Severity eq 'Critical'\").",
	}, c.listAlarms)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_hcl_statuses",
		Description: "List Hardware Compatibility List (HCL) validation statuses. Shows compliance status, hardware/software validation, and reasons for non-compliance. Supports OData $filter.",
	}, c.listHclStatuses)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_firmware",
		Description: "List running firmware versions across server components. Returns component name, version, and type. Supports OData $filter.",
	}, c.listFirmware)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_organizations",
		Description: "List organizations in the Intersight account. Returns name and description.",
	}, c.listOrganizations)
}

func (c *Client) listServers(ctx context.Context, req *mcp.CallToolRequest, args FilterArgs) (*mcp.CallToolResult, any, error) {
	r := c.API.ComputeApi.GetComputePhysicalSummaryList(c.Ctx)
	if args.Filter != "" {
		r = r.Filter(args.Filter)
	}
	if args.Top > 0 {
		r = r.Top(args.Top)
	}
	resp, _, err := r.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("list servers: %w", err)
	}
	if resp.ComputePhysicalSummaryList == nil {
		return textResult("No servers found."), nil, nil
	}
	return textResult(formatServers(resp.ComputePhysicalSummaryList.GetResults())), nil, nil
}

func (c *Client) listAlarms(ctx context.Context, req *mcp.CallToolRequest, args FilterArgs) (*mcp.CallToolResult, any, error) {
	r := c.API.CondApi.GetCondAlarmList(c.Ctx)
	if args.Filter != "" {
		r = r.Filter(args.Filter)
	}
	if args.Top > 0 {
		r = r.Top(args.Top)
	}
	resp, _, err := r.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("list alarms: %w", err)
	}
	if resp.CondAlarmList == nil {
		return textResult("No alarms found."), nil, nil
	}
	return textResult(formatAlarms(resp.CondAlarmList.GetResults())), nil, nil
}

func (c *Client) listHclStatuses(ctx context.Context, req *mcp.CallToolRequest, args FilterArgs) (*mcp.CallToolResult, any, error) {
	r := c.API.CondApi.GetCondHclStatusList(c.Ctx)
	if args.Filter != "" {
		r = r.Filter(args.Filter)
	}
	if args.Top > 0 {
		r = r.Top(args.Top)
	}
	resp, _, err := r.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("list HCL statuses: %w", err)
	}
	if resp.CondHclStatusList == nil {
		return textResult("No HCL statuses found."), nil, nil
	}
	return textResult(formatHclStatuses(resp.CondHclStatusList.GetResults())), nil, nil
}

func (c *Client) listFirmware(ctx context.Context, req *mcp.CallToolRequest, args FilterArgs) (*mcp.CallToolResult, any, error) {
	r := c.API.FirmwareApi.GetFirmwareRunningFirmwareList(c.Ctx)
	if args.Filter != "" {
		r = r.Filter(args.Filter)
	}
	if args.Top > 0 {
		r = r.Top(args.Top)
	}
	resp, _, err := r.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("list firmware: %w", err)
	}
	if resp.FirmwareRunningFirmwareList == nil {
		return textResult("No firmware entries found."), nil, nil
	}
	return textResult(formatFirmware(resp.FirmwareRunningFirmwareList.GetResults())), nil, nil
}

func (c *Client) listOrganizations(ctx context.Context, req *mcp.CallToolRequest, args NoArgs) (*mcp.CallToolResult, any, error) {
	resp, _, err := c.API.OrganizationApi.GetOrganizationOrganizationList(c.Ctx).Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("list organizations: %w", err)
	}
	if resp.OrganizationOrganizationList == nil {
		return textResult("No organizations found."), nil, nil
	}
	return textResult(formatOrganizations(resp.OrganizationOrganizationList.GetResults())), nil, nil
}
