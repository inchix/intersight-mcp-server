package intersight

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/CiscoDevNet/intersight-go"
)

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"intersight.com", "intersight.com"},
		{"  intersight.com  ", "intersight.com"},
		{"https://intersight.com", "intersight.com"},
		{"http://intersight.com", "intersight.com"},
		{"https://intersight.com/", "intersight.com"},
		{"appliance.example.com/", "appliance.example.com"},
	}
	for _, tt := range tests {
		if got := normalizeHost(tt.in); got != tt.want {
			t.Errorf("normalizeHost(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// fakeRequest mirrors the generated SDK builders: a value type whose option
// methods return a modified copy.
type fakeRequest struct {
	filter, orderby, inlinecount string
	top                          int32
}

func (f fakeRequest) Filter(s string) fakeRequest      { f.filter = s; return f }
func (f fakeRequest) Orderby(s string) fakeRequest     { f.orderby = s; return f }
func (f fakeRequest) Top(t int32) fakeRequest          { f.top = t; return f }
func (f fakeRequest) Inlinecount(s string) fakeRequest { f.inlinecount = s; return f }

func TestApplyArgs(t *testing.T) {
	tests := []struct {
		name string
		args FilterArgs
		want fakeRequest
	}{
		{
			name: "empty args set only inlinecount",
			args: FilterArgs{},
			want: fakeRequest{inlinecount: "allpages"},
		},
		{
			name: "all options forwarded",
			args: FilterArgs{Filter: "Severity eq 'Critical'", OrderBy: "CreationTime desc", Top: 50},
			want: fakeRequest{filter: "Severity eq 'Critical'", orderby: "CreationTime desc", top: 50, inlinecount: "allpages"},
		},
		{
			name: "top clamped to API maximum",
			args: FilterArgs{Top: 5000},
			want: fakeRequest{top: maxTop, inlinecount: "allpages"},
		},
		{
			name: "top at the maximum is unchanged",
			args: FilterArgs{Top: maxTop},
			want: fakeRequest{top: maxTop, inlinecount: "allpages"},
		},
		{
			name: "negative top is not forwarded",
			args: FilterArgs{Top: -1},
			want: fakeRequest{inlinecount: "allpages"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applyArgs(fakeRequest{}, tt.args); got != tt.want {
				t.Errorf("applyArgs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestHeading(t *testing.T) {
	t.Run("complete result set", func(t *testing.T) {
		var b strings.Builder
		heading(&b, "server(s)", 3, 3)
		if got := b.String(); !strings.HasPrefix(got, "Found 3 server(s):") {
			t.Errorf("got %q, want a \"Found 3\" heading", got)
		}
	})

	t.Run("truncated result set reports the total", func(t *testing.T) {
		var b strings.Builder
		heading(&b, "server(s)", 100, 1543)
		got := b.String()
		if !strings.Contains(got, "Showing 100 of 1543 server(s)") {
			t.Errorf("got %q, want it to report 100 of 1543", got)
		}
		if !strings.Contains(got, "truncated") {
			t.Errorf("got %q, want it to flag truncation", got)
		}
	})

	t.Run("absent inline count does not claim truncation", func(t *testing.T) {
		// If Intersight omits $inlinecount, GetCount returns 0; that must not be
		// read as "0 total, 5 shown".
		var b strings.Builder
		heading(&b, "server(s)", 5, 0)
		if got := b.String(); !strings.HasPrefix(got, "Found 5 server(s):") {
			t.Errorf("got %q, want a plain \"Found 5\" heading", got)
		}
	})
}

func TestFormattersOnEmptyInput(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"servers", formatServers(nil, 0), "No servers found."},
		{"alarms", formatAlarms(nil, 0), "No alarms found."},
		{"hcl", formatHclStatuses(nil, 0), "No HCL statuses found."},
		{"firmware", formatFirmware(nil, 0), "No firmware entries found."},
		{"organizations", formatOrganizations(nil, 0), "No organizations found."},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestAPIErrorWrapsUnderlyingError(t *testing.T) {
	sentinel := errors.New("connection refused")
	err := apiError("list servers", fmt.Errorf("get: %w", sentinel))

	if !errors.Is(err, sentinel) {
		t.Errorf("apiError should preserve the error chain, got %v", err)
	}
	if !strings.HasPrefix(err.Error(), "list servers: ") {
		t.Errorf("apiError should prefix the operation, got %q", err.Error())
	}
}

// TestAPIErrorSurfacesResponseBody drives a real SDK call against a stub server
// so the error travels the same path it does in production. The SDK returns
// *GenericOpenAPIError rather than a value, and matching the wrong one silently
// drops Intersight's explanation -- the whole point of apiError.
func TestAPIErrorSurfacesResponseBody(t *testing.T) {
	const body = `{"code":"InvalidUrl","message":"$filter is malformed"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, body)
	}))
	defer srv.Close()

	cfg := sdk.NewConfiguration()
	cfg.Servers = sdk.ServerConfigurations{{URL: srv.URL}}
	api := sdk.NewAPIClient(cfg)

	_, _, err := api.ComputeApi.GetComputePhysicalSummaryList(context.Background()).
		Filter("bogus").Execute()
	if err == nil {
		t.Fatal("expected the stub server's 400 to produce an error")
	}

	got := apiError("list servers", err).Error()
	if !strings.HasPrefix(got, "list servers: ") {
		t.Errorf("missing operation prefix: %q", got)
	}
	if !strings.Contains(got, "$filter is malformed") {
		t.Errorf("apiError dropped the Intersight response body; got %q", got)
	}
}

// TestAPIErrorTruncatesHugeBody guards the context window against an oversized
// error payload.
func TestAPIErrorTruncatesHugeBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, strings.Repeat("x", maxErrBody*3))
	}))
	defer srv.Close()

	cfg := sdk.NewConfiguration()
	cfg.Servers = sdk.ServerConfigurations{{URL: srv.URL}}
	api := sdk.NewAPIClient(cfg)

	_, _, err := api.ComputeApi.GetComputePhysicalSummaryList(context.Background()).Execute()
	if err == nil {
		t.Fatal("expected the stub server's 500 to produce an error")
	}

	got := apiError("list servers", err).Error()
	if len(got) > maxErrBody+256 {
		t.Errorf("error not truncated: length %d", len(got))
	}
	if !strings.Contains(got, "(truncated)") {
		t.Errorf("expected a truncation marker, got %q", got[:200])
	}
}
