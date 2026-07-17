-- +goose Up
CREATE TABLE sso_connections (
    id                    uuid PRIMARY KEY DEFAULT uuidv7(),
    workspace_id          uuid NOT NULL UNIQUE REFERENCES workspaces(id) ON DELETE CASCADE,
    mode                  text NOT NULL,
    issuer                text NOT NULL DEFAULT '',
    client_id             text NOT NULL DEFAULT '',
    client_secret_enc     bytea,
    scopes                text[] NOT NULL DEFAULT ARRAY['openid','email','profile']::text[],
    saml_metadata_url     text NOT NULL DEFAULT '',
    enabled               boolean NOT NULL DEFAULT false,
    required              boolean NOT NULL DEFAULT false,
    jit_provisioning      boolean NOT NULL DEFAULT false,
    allowed_email_domains text[] NOT NULL DEFAULT ARRAY[]::text[],
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT sso_connections_mode CHECK (mode IN ('oidc','saml'))
);

-- +goose Down
DROP TABLE sso_connections;
