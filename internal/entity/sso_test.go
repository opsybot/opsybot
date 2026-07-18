package entity

import (
	"slices"
	"testing"
)

func TestEmailDomainAllowed(t *testing.T) {
	cases := []struct {
		name    string
		email   string
		domains []string
		want    bool
	}{
		{"no restriction allows any", "a@anywhere.io", nil, true},
		{"exact domain match", "a@acme.dev", []string{"acme.dev"}, true},
		{"case-insensitive match", "a@ACME.dev", []string{"acme.dev"}, true},
		{"non-matching domain denied", "a@evil.com", []string{"acme.dev"}, false},
		{"one of several", "a@corp.io", []string{"acme.dev", "corp.io"}, true},
		{"missing @ denied when restricted", "not-an-email", []string{"acme.dev"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := EmailDomainAllowed(c.email, c.domains); got != c.want {
				t.Errorf("EmailDomainAllowed(%q, %v) = %v, want %v", c.email, c.domains, got, c.want)
			}
		})
	}
}

func TestNormalizeDomains(t *testing.T) {
	got := NormalizeDomains([]string{"Acme.dev", " acme.dev ", "", "CORP.io", "corp.io"})
	want := []string{"acme.dev", "corp.io"}
	if !slices.Equal(got, want) {
		t.Errorf("NormalizeDomains = %v, want %v", got, want)
	}
}

func TestSSOConfigInputValidate(t *testing.T) {
	cases := []struct {
		name    string
		in      SSOConfigInput
		wantErr bool
	}{
		{"valid oidc", SSOConfigInput{Mode: SSOModeOIDC, Issuer: "https://idp.example.com", ClientID: "app"}, false},
		{"oidc missing issuer", SSOConfigInput{Mode: SSOModeOIDC, ClientID: "app"}, true},
		{"oidc missing client id", SSOConfigInput{Mode: SSOModeOIDC, Issuer: "https://idp.example.com"}, true},
		{"oidc bad issuer url", SSOConfigInput{Mode: SSOModeOIDC, Issuer: "not a url", ClientID: "app"}, true},
		{"valid saml", SSOConfigInput{Mode: SSOModeSAML, SAMLMetadataURL: "https://idp.example.com/metadata"}, false},
		{"saml missing metadata", SSOConfigInput{Mode: SSOModeSAML}, true},
		{"bad mode", SSOConfigInput{Mode: SSOMode("ldap")}, true},
		{"domain with at-sign", SSOConfigInput{Mode: SSOModeOIDC, Issuer: "https://idp.example.com", ClientID: "app", AllowedEmailDomains: []string{"a@b"}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.in.Validate()
			if (err != nil) != c.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}
