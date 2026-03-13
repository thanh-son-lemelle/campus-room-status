package adminsdk

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestNewServiceAccountTokenProvider_RequiresCredentials(t *testing.T) {
	t.Parallel()

	_, err := NewServiceAccountTokenProvider(ServiceAccountTokenProviderConfig{})
	if err == nil {
		t.Fatalf("expected error when credentials are missing")
	}
}

func TestNewServiceAccountTokenProvider_RejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := NewServiceAccountTokenProvider(ServiceAccountTokenProviderConfig{
		CredentialsJSON: []byte(`{"type":"service_account","private_key":"invalid"`),
	})
	if err == nil {
		t.Fatalf("expected parse error for invalid credentials JSON")
	}
}

func TestServiceAccountTokenProvider_TokenReturnsAccessToken(t *testing.T) {
	t.Parallel()

	provider, err := NewServiceAccountTokenProvider(ServiceAccountTokenProviderConfig{
		CredentialsJSON: makeServiceAccountJSON(t),
		Subject:         "admin@example.org",
	})
	if err != nil {
		t.Fatalf("expected provider creation to succeed, got %v", err)
	}

	impl, ok := provider.(*serviceAccountTokenProvider)
	if !ok {
		t.Fatalf("expected concrete provider type, got %T", provider)
	}

	impl.sourceFactory = func(context.Context) oauth2.TokenSource {
		return tokenSourceFunc(func() (*oauth2.Token, error) {
			return &oauth2.Token{
				AccessToken: "service-account-token",
			}, nil
		})
	}

	token, err := provider.Token(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "service-account-token" {
		t.Fatalf("expected service-account-token, got %q", token)
	}
}

func TestServiceAccountTokenProvider_TokenReturnsErrorWhenTokenSourceFails(t *testing.T) {
	t.Parallel()

	provider := &serviceAccountTokenProvider{
		sourceFactory: func(context.Context) oauth2.TokenSource {
			return tokenSourceFunc(func() (*oauth2.Token, error) {
				return nil, errors.New("token source unavailable")
			})
		},
	}

	_, err := provider.Token(context.Background())
	if err == nil {
		t.Fatalf("expected token retrieval error")
	}
	if !strings.Contains(err.Error(), "token source unavailable") {
		t.Fatalf("expected wrapped token source error, got %v", err)
	}
}

func TestServiceAccountTokenProvider_TokenReturnsErrorOnEmptyAccessToken(t *testing.T) {
	t.Parallel()

	provider := &serviceAccountTokenProvider{
		sourceFactory: func(context.Context) oauth2.TokenSource {
			return tokenSourceFunc(func() (*oauth2.Token, error) {
				return &oauth2.Token{AccessToken: "   "}, nil
			})
		},
	}

	_, err := provider.Token(context.Background())
	if err == nil {
		t.Fatalf("expected empty token error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "empty") {
		t.Fatalf("expected empty token error, got %v", err)
	}
}

func makeServiceAccountJSON(t *testing.T) []byte {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyDER,
	})

	payload := map[string]any{
		"type":                        "service_account",
		"project_id":                  "test-project",
		"private_key_id":              "test-private-key-id",
		"private_key":                 string(privateKeyPEM),
		"client_email":                "service-account@test-project.iam.gserviceaccount.com",
		"client_id":                   "123456789012345678901",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/service-account%40test-project.iam.gserviceaccount.com",
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal service account JSON: %v", err)
	}
	return raw
}

type tokenSourceFunc func() (*oauth2.Token, error)

func (f tokenSourceFunc) Token() (*oauth2.Token, error) {
	return f()
}
