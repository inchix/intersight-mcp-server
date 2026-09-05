package intersight

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	intersight "github.com/CiscoDevNet/intersight-go"
)

// Version is the server version, reported in the Intersight User-Agent header.
// It is overridden from main at startup.
var Version = "dev"

// requestTimeout bounds a single Intersight API call. The SDK otherwise falls
// back to http.DefaultClient, which has no timeout at all.
const requestTimeout = 60 * time.Second

// Client holds an authenticated Intersight API client.
type Client struct {
	API *intersight.APIClient

	// auth is the validated HTTP-signature credential, with the private key
	// already loaded. Context returns a context carrying it.
	auth any
}

// NewClient creates a new authenticated Intersight client.
// host is optional; defaults to "intersight.com".
func NewClient(keyID, keyFile, host string) (*Client, error) {
	cfg := intersight.NewConfiguration()
	cfg.UserAgent = "intersight-mcp-server/" + Version
	cfg.HTTPClient = &http.Client{Timeout: requestTimeout}

	if host = normalizeHost(host); host != "" {
		cfg.Servers = intersight.ServerConfigurations{
			{URL: "https://" + host},
		}
	}

	authCfg := intersight.HttpSignatureAuth{
		KeyId:            keyID,
		PrivateKeyPath:   keyFile,
		SigningScheme:    "hs2019",
		SigningAlgorithm: "RSASSA-PKCS1-v1_5",
		HashAlgorithm:    "sha256",
		SignedHeaders:    []string{"(request-target)", "Host", "Date", "Digest"},
	}

	// ContextWithValue validates the credential and reads the private key from
	// disk. We do that once here, then keep only the resulting value so each
	// request can carry its own cancellable context rather than a stored one.
	authCtx, err := authCfg.ContextWithValue(context.Background())
	if err != nil {
		return nil, fmt.Errorf("intersight auth: %w", err)
	}
	auth := authCtx.Value(intersight.ContextHttpSignatureAuth)
	if auth == nil {
		return nil, fmt.Errorf("intersight auth: credential missing from context")
	}

	return &Client{
		API:  intersight.NewAPIClient(cfg),
		auth: auth,
	}, nil
}

// Context returns ctx carrying the Intersight credential, so that API calls
// inherit the caller's deadline and cancellation.
func (c *Client) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, intersight.ContextHttpSignatureAuth, c.auth)
}

// normalizeHost trims whitespace, any scheme, and trailing slashes, so that
// INTERSIGHT_API_HOST accepts "intersight.com" as well as
// "https://intersight.com/" without producing a "https://https://..." URL.
func normalizeHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	return strings.Trim(host, "/")
}
