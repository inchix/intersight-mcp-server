package intersight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	intersight "github.com/CiscoDevNet/intersight-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxTop is the largest $top the Intersight API accepts; larger values are
// rejected server-side, so we clamp instead of forwarding them.
const maxTop = 1000

// maxErrBody caps how much of an Intersight error payload we echo back.
const maxErrBody = 2048

// FilterArgs is the shared input schema for tools that support OData filtering.
type FilterArgs struct {
	Filter  string `json:"filter,omitempty" jsonschema:"OData $filter expression, e.g. \"Model eq 'UCSC-C220-M5SX'\" or \"Severity eq 'Critical'\". String literals must be single-quoted."`
	OrderBy string `json:"orderby,omitempty" jsonschema:"OData $orderby expression, e.g. \"Name\" or \"CreationTime desc\". Comma-separate multiple properties."`
	Top     int32  `json:"top,omitzero" jsonschema:"Maximum number of results to return. Defaults to 100 if omitted; values above 1000 are clamped to 1000."`
}

// readOnly marks a tool as making no changes to the Intersight account, which
// lets clients skip write-approval prompts.
func readOnly() *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true}
}

// RegisterTools adds all Intersight MCP tools to the server.
func RegisterTools(s *mcp.Server, c *Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_servers",
		Description: "List physical server inventory from Intersight. Returns name, model, serial, power state, CPU/memory, management IP, and firmware. Supports OData $filter and $orderby.",
		Annotations: readOnly(),
	}, c.listServers)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_alarms",
		Description: "List active alarms from Intersight. Returns severity, description, affected object, and creation time. Supports OData $filter and $orderby (e.g. filter \"Severity eq 'Critical'\").",
		Annotations: readOnly(),
	}, c.listAlarms)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_hcl_statuses",
		Description: "List Hardware Compatibility List (HCL) validation statuses. Shows compliance status, hardware/software validation, and reasons for non-compliance. Supports OData $filter and $orderby.",
		Annotations: readOnly(),
	}, c.listHclStatuses)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_firmware",
		Description: "List running firmware versions across server components. Returns component name, version, and type. Supports OData $filter and $orderby.",
		Annotations: readOnly(),
	}, c.listFirmware)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_organizations",
		Description: "List organizations in the Intersight account. Returns name and description. Supports OData $filter and $orderby.",
		Annotations: readOnly(),
	}, c.listOrganizations)
}

// listRequest is the subset of the generated per-endpoint request builders that
// every list tool uses. The builders are value types whose option methods
// return a modified copy, hence the self-referential type parameter.
type listRequest[T any] interface {
	Filter(string) T
	Orderby(string) T
	Top(int32) T
	Inlinecount(string) T
}

// applyArgs applies the shared OData options to a list request. It always asks
// for an inline count so the caller can tell a truncated page from a complete
// result set.
func applyArgs[T listRequest[T]](r T, args FilterArgs) T {
	if args.Filter != "" {
		r = r.Filter(args.Filter)
	}
	if args.OrderBy != "" {
		r = r.Orderby(args.OrderBy)
	}
	if args.Top > 0 {
		r = r.Top(min(args.Top, maxTop))
	}
	return r.Inlinecount("allpages")
}

// apiError unwraps the SDK's generic error so the caller sees Intersight's own
// explanation (a malformed $filter, an expired key) instead of a bare status
// line, which is what lets a model correct its own query.
func apiError(op string, err error) error {
	// The generated SDK always returns *GenericOpenAPIError, never a value, so
	// the errors.As target must be the pointer type.
	var genErr *intersight.GenericOpenAPIError
	if errors.As(err, &genErr) {
		if body := strings.TrimSpace(string(genErr.Body())); body != "" {
			if len(body) > maxErrBody {
				body = body[:maxErrBody] + "... (truncated)"
			}
			return fmt.Errorf("%s: %s: %s", op, genErr.Error(), body)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}

func (c *Client) listServers(ctx context.Context, req *mcp.CallToolRequest, args FilterArgs) (*mcp.CallToolResult, any, error) {
	r := applyArgs(c.API.ComputeApi.GetComputePhysicalSummaryList(c.Context(ctx)), args)
	resp, _, err := r.Execute()
	if err != nil {
		return nil, nil, apiError("list servers", err)
	}
	list := resp.ComputePhysicalSummaryList
	if list == nil {
		return textResult("No servers found."), nil, nil
	}
	return textResult(formatServers(list.GetResults(), list.GetCount())), nil, nil
}

func (c *Client) listAlarms(ctx context.Context, req *mcp.CallToolRequest, args FilterArgs) (*mcp.CallToolResult, any, error) {
	r := applyArgs(c.API.CondApi.GetCondAlarmList(c.Context(ctx)), args)
	resp, _, err := r.Execute()
	if err != nil {
		return nil, nil, apiError("list alarms", err)
	}
	list := resp.CondAlarmList
	if list == nil {
		return textResult("No alarms found."), nil, nil
	}
	return textResult(formatAlarms(list.GetResults(), list.GetCount())), nil, nil
}

func (c *Client) listHclStatuses(ctx context.Context, req *mcp.CallToolRequest, args FilterArgs) (*mcp.CallToolResult, any, error) {
	r := applyArgs(c.API.CondApi.GetCondHclStatusList(c.Context(ctx)), args)
	resp, _, err := r.Execute()
	if err != nil {
		return nil, nil, apiError("list HCL statuses", err)
	}
	list := resp.CondHclStatusList
	if list == nil {
		return textResult("No HCL statuses found."), nil, nil
	}
	return textResult(formatHclStatuses(list.GetResults(), list.GetCount())), nil, nil
}

func (c *Client) listFirmware(ctx context.Context, req *mcp.CallToolRequest, args FilterArgs) (*mcp.CallToolResult, any, error) {
	r := applyArgs(c.API.FirmwareApi.GetFirmwareRunningFirmwareList(c.Context(ctx)), args)
	resp, _, err := r.Execute()
	if err != nil {
		return nil, nil, apiError("list firmware", err)
	}
	list := resp.FirmwareRunningFirmwareList
	if list == nil {
		return textResult("No firmware entries found."), nil, nil
	}
	return textResult(formatFirmware(list.GetResults(), list.GetCount())), nil, nil
}

func (c *Client) listOrganizations(ctx context.Context, req *mcp.CallToolRequest, args FilterArgs) (*mcp.CallToolResult, any, error) {
	r := applyArgs(c.API.OrganizationApi.GetOrganizationOrganizationList(c.Context(ctx)), args)
	resp, _, err := r.Execute()
	if err != nil {
		return nil, nil, apiError("list organizations", err)
	}
	list := resp.OrganizationOrganizationList
	if list == nil {
		return textResult("No organizations found."), nil, nil
	}
	return textResult(formatOrganizations(list.GetResults(), list.GetCount())), nil, nil
}
