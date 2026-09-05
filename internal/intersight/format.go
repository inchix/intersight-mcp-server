package intersight

import (
	"fmt"
	"strings"

	intersight "github.com/CiscoDevNet/intersight-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

// heading opens a result listing. total comes from the API's $inlinecount and
// counts every match, not just this page, so a shown < total tells the caller
// the result was truncated and how much it is missing.
func heading(b *strings.Builder, noun string, shown int, total int32) {
	if int64(total) > int64(shown) {
		fmt.Fprintf(b, "Showing %d of %d %s (result truncated; raise `top` up to %d or narrow `filter`):\n\n",
			shown, total, noun, maxTop)
		return
	}
	fmt.Fprintf(b, "Found %d %s:\n\n", shown, noun)
}

func formatServers(servers []intersight.ComputePhysicalSummary, total int32) string {
	if len(servers) == 0 {
		return "No servers found."
	}
	var b strings.Builder
	heading(&b, "server(s)", len(servers), total)
	for _, s := range servers {
		fmt.Fprintf(&b, "- %s\n", s.GetName())
		fmt.Fprintf(&b, "  Model: %s | Serial: %s\n", s.GetModel(), s.GetSerial())
		fmt.Fprintf(&b, "  Power: %s | State: %s\n", s.GetOperPowerState(), s.GetOperState())
		fmt.Fprintf(&b, "  CPUs: %d | Cores: %d | Memory: %d MB\n",
			s.GetNumCpus(), s.GetNumCpuCores(), s.GetTotalMemory())
		fmt.Fprintf(&b, "  Mgmt IP: %s | Firmware: %s\n", s.GetMgmtIpAddress(), s.GetFirmware())
		b.WriteString("\n")
	}
	return b.String()
}

func formatAlarms(alarms []intersight.CondAlarm, total int32) string {
	if len(alarms) == 0 {
		return "No alarms found."
	}
	var b strings.Builder
	heading(&b, "alarm(s)", len(alarms), total)
	for _, a := range alarms {
		fmt.Fprintf(&b, "- [%s] %s\n", a.GetSeverity(), a.GetName())
		fmt.Fprintf(&b, "  Description: %s\n", a.GetDescription())
		fmt.Fprintf(&b, "  Affected: %s (%s)\n", a.GetAffectedMoDisplayName(), a.GetAffectedMoType())
		fmt.Fprintf(&b, "  Created: %s\n", a.GetCreationTime())
		b.WriteString("\n")
	}
	return b.String()
}

func formatHclStatuses(statuses []intersight.CondHclStatus, total int32) string {
	if len(statuses) == 0 {
		return "No HCL statuses found."
	}
	var b strings.Builder
	heading(&b, "HCL status(es)", len(statuses), total)
	for _, h := range statuses {
		fmt.Fprintf(&b, "- %s: %s\n", h.GetServerName(), h.GetStatus())
		fmt.Fprintf(&b, "  Hardware: %s | Software: %s\n",
			h.GetHardwareStatus(), h.GetSoftwareStatus())
		if reason := h.GetReason(); reason != "" {
			fmt.Fprintf(&b, "  Reason: %s\n", reason)
		}
		fmt.Fprintf(&b, "  Model: %s | Firmware: %s | OS: %s %s\n",
			h.GetHclModel(), h.GetHclFirmwareVersion(),
			h.GetHclOsVendor(), h.GetHclOsVersion())
		b.WriteString("\n")
	}
	return b.String()
}

func formatFirmware(firmwares []intersight.FirmwareRunningFirmware, total int32) string {
	if len(firmwares) == 0 {
		return "No firmware entries found."
	}
	var b strings.Builder
	heading(&b, "firmware entry(ies)", len(firmwares), total)
	for _, f := range firmwares {
		fmt.Fprintf(&b, "- %s: %s (type: %s)\n",
			f.GetComponent(), f.GetVersion(), f.GetType())
		if pkg := f.GetPackageVersion(); pkg != "" {
			fmt.Fprintf(&b, "  Package: %s\n", pkg)
		}
	}
	return b.String()
}

func formatOrganizations(orgs []intersight.OrganizationOrganization, total int32) string {
	if len(orgs) == 0 {
		return "No organizations found."
	}
	var b strings.Builder
	heading(&b, "organization(s)", len(orgs), total)
	for _, o := range orgs {
		desc := o.GetDescription()
		if desc == "" {
			desc = "(no description)"
		}
		fmt.Fprintf(&b, "- %s: %s\n", o.GetName(), desc)
	}
	return b.String()
}
