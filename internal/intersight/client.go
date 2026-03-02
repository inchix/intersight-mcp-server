package intersight

import (
	"context"
	"fmt"

	intersight "github.com/CiscoDevNet/intersight-go"
)

// Client holds an authenticated Intersight API client and context.
type Client struct {
	API *intersight.APIClient
	Ctx context.Context
}

// NewClient creates a new authenticated Intersight client.
// host is optional; defaults to "intersight.com".
func NewClient(keyID, keyFile, host string) (*Client, error) {
	cfg := intersight.NewConfiguration()

	if host != "" {
		cfg.Servers = intersight.ServerConfigurations{
			{URL: "https://" + host},
		}
	}

	authCfg := intersight.HttpSignatureAuth{
		KeyId:            keyID,
		PrivateKeyPath:   keyFile,
		SigningScheme:    "hs2019",
		SigningAlgorithm: "RSASSA-PKCS1-v1_5",
		HashAlgorithm:   "sha256",
		SignedHeaders:    []string{"(request-target)", "Host", "Date", "Digest"},
	}

	ctx, err := authCfg.ContextWithValue(context.Background())
	if err != nil {
		return nil, fmt.Errorf("intersight auth: %w", err)
	}

	return &Client{
		API: intersight.NewAPIClient(cfg),
		Ctx: ctx,
	}, nil
}
