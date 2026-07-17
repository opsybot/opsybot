package sso

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"

	"github.com/opsybot/opsybot/internal/entity"
)

func (s *srv) startSAML(ctx context.Context, ws entity.Workspace, conn entity.SSOConnection, workspaceSlug string) (string, error) {
	sp, err := s.serviceProvider(ctx, conn, workspaceSlug)
	if err != nil {
		return "", err
	}
	relayState, err := entity.GenerateToken(entity.SSOStateTokenLen)
	if err != nil {
		return "", err
	}
	authnReq, err := sp.MakeAuthenticationRequest(sp.GetSSOBindingLocation(saml.HTTPRedirectBinding), saml.HTTPRedirectBinding, saml.HTTPPostBinding)
	if err != nil {
		return "", fmt.Errorf("%w: %v", entity.ErrSSOInvalid, err)
	}
	if err := s.states.Store(ctx, relayState, entity.SSOState{
		WorkspaceID: ws.ID, ConnectionID: conn.ID, Nonce: authnReq.ID,
	}, entity.SSOStateTTL); err != nil {
		return "", err
	}
	redirect, err := authnReq.Redirect(relayState, &sp)
	if err != nil {
		return "", fmt.Errorf("%w: %v", entity.ErrSSOInvalid, err)
	}
	return redirect.String(), nil
}

func (s *srv) CompleteSAML(ctx context.Context, workspaceSlug, samlResponse, relayState, ip, userAgent string) (entity.LoginResult, error) {
	ws, err := s.workspaces.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return entity.LoginResult{}, err
	}
	st, err := s.states.Consume(ctx, relayState)
	if err != nil {
		return entity.LoginResult{}, err
	}
	if st.WorkspaceID != ws.ID {
		return entity.LoginResult{}, entity.ErrSSOStateInvalid
	}
	conn, err := s.connections.Get(ctx, ws.ID)
	if err != nil {
		return entity.LoginResult{}, err
	}
	if conn.Mode != entity.SSOModeSAML || conn.ID != st.ConnectionID {
		return entity.LoginResult{}, entity.ErrSSOStateInvalid
	}
	sp, err := s.serviceProvider(ctx, conn, workspaceSlug)
	if err != nil {
		return entity.LoginResult{}, err
	}
	raw, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return entity.LoginResult{}, entity.ErrSSOExchange
	}
	assertion, err := sp.ParseXMLResponse(raw, []string{st.Nonce}, sp.AcsURL)
	if err != nil {
		return entity.LoginResult{}, fmt.Errorf("%w: %v", entity.ErrSSOExchange, err)
	}

	email := entity.NormalizeEmail(samlAttr(assertion, "email"))
	subject := samlNameID(assertion)
	if email == "" {
		email = entity.NormalizeEmail(subject)
	}
	if email == "" {
		return entity.LoginResult{}, entity.ErrSSOEmailMissing
	}
	if subject == "" {
		subject = email
	}
	name := strings.TrimSpace(samlAttr(assertion, "name"))
	if name == "" {
		name = email
	}
	return s.finishLogin(ctx, ws, conn, subject, email, name, ip, userAgent)
}

func (s *srv) SAMLMetadata(ctx context.Context, workspaceSlug string) ([]byte, error) {
	ws, err := s.workspaces.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return nil, err
	}
	conn, err := s.connections.Get(ctx, ws.ID)
	if err != nil {
		return nil, err
	}
	if conn.Mode != entity.SSOModeSAML {
		return nil, entity.ErrSSOInvalid
	}
	sp, err := s.serviceProvider(ctx, conn, workspaceSlug)
	if err != nil {
		return nil, err
	}
	return xml.MarshalIndent(sp.Metadata(), "", "  ")
}

func (s *srv) serviceProvider(ctx context.Context, conn entity.SSOConnection, workspaceSlug string) (saml.ServiceProvider, error) {
	idpMeta, err := s.samlMetadata(ctx, conn.SAMLMetadataURL)
	if err != nil {
		return saml.ServiceProvider{}, err
	}
	key, cert, err := s.samlKeypair()
	if err != nil {
		return saml.ServiceProvider{}, err
	}
	base := s.samlBase(workspaceSlug)
	metaURL, err := url.Parse(base + "/metadata")
	if err != nil {
		return saml.ServiceProvider{}, entity.ErrSSOInvalid
	}
	acsURL, err := url.Parse(base + "/acs")
	if err != nil {
		return saml.ServiceProvider{}, entity.ErrSSOInvalid
	}
	return saml.ServiceProvider{
		EntityID:          metaURL.String(),
		Key:               key,
		Certificate:       cert,
		MetadataURL:       *metaURL,
		AcsURL:            *acsURL,
		IDPMetadata:       idpMeta,
		AuthnNameIDFormat: saml.EmailAddressNameIDFormat,
		AllowIDPInitiated: false,
	}, nil
}

func (s *srv) samlBase(workspaceSlug string) string {
	return strings.TrimRight(s.cfg.BaseURL, "/") + "/v1/auth/sso/" + workspaceSlug + "/saml"
}

func (s *srv) samlMetadata(ctx context.Context, metadataURL string) (*saml.EntityDescriptor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.samlMeta[metadataURL]; ok {
		return m, nil
	}
	u, err := url.Parse(metadataURL)
	if err != nil {
		return nil, entity.ErrSSOInvalid
	}
	m, err := samlsp.FetchMetadata(ctx, http.DefaultClient, *u)
	if err != nil {
		return nil, fmt.Errorf("saml metadata: %w", err)
	}
	s.samlMeta[metadataURL] = m
	return m, nil
}

func (s *srv) samlKeypair() (*rsa.PrivateKey, *x509.Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.spKey != nil {
		return s.spKey.(*rsa.PrivateKey), s.spCert, nil
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate sp key: %w", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "opsybot-sp"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(entity.SSOSPCertValidity),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, fmt.Errorf("create sp cert: %w", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, fmt.Errorf("parse sp cert: %w", err)
	}
	s.spKey = key
	s.spCert = cert
	return key, cert, nil
}

func samlAttr(a *saml.Assertion, name string) string {
	if a == nil {
		return ""
	}
	for _, stmt := range a.AttributeStatements {
		for _, attr := range stmt.Attributes {
			if attr.Name != name && attr.FriendlyName != name {
				continue
			}
			for _, v := range attr.Values {
				if v.Value != "" {
					return v.Value
				}
			}
		}
	}
	return ""
}

func samlNameID(a *saml.Assertion) string {
	if a == nil || a.Subject == nil || a.Subject.NameID == nil {
		return ""
	}
	return a.Subject.NameID.Value
}
