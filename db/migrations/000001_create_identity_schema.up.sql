
-- HIREOS identity and organization foundation.
-- Migration: 000001

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CHECK (length(btrim(name)) > 0),
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL
        CHECK (email = lower(btrim(email))),
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL
        CHECK (length(btrim(display_name)) > 0),
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'disabled')),
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL
        REFERENCES organizations(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'removed')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT organization_members_unique
        UNIQUE (organization_id, user_id),

    CONSTRAINT organization_members_id_org_unique
        UNIQUE (id, organization_id)
);

CREATE INDEX organization_members_user_id_idx
    ON organization_members(user_id);

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_key TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permission_key TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL
        REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL
        REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE member_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    organization_member_id UUID NOT NULL,
    role_id UUID NOT NULL
        REFERENCES roles(id) ON DELETE RESTRICT,
    assigned_by_user_id UUID
        REFERENCES users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT member_roles_membership_fk
        FOREIGN KEY (organization_member_id, organization_id)
        REFERENCES organization_members(id, organization_id)
        ON DELETE CASCADE,

    CONSTRAINT member_roles_unique
        UNIQUE (organization_member_id, role_id)
);

CREATE INDEX member_roles_role_id_idx
    ON member_roles(role_id);

CREATE TABLE candidate_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE
        REFERENCES users(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL
        REFERENCES organizations(id) ON DELETE RESTRICT,
    email TEXT NOT NULL
        CHECK (email = lower(btrim(email))),
    token_hash TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (
            status IN ('pending', 'accepted', 'revoked', 'expired')
        ),
    invited_by_user_id UUID
        REFERENCES users(id) ON DELETE SET NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX invitations_organization_id_idx
    ON invitations(organization_id);

CREATE UNIQUE INDEX invitations_pending_email_unique
    ON invitations(organization_id, email)
    WHERE status = 'pending';

-- Initial role catalogue from the HIREOS PRD.
INSERT INTO roles (role_key, name) VALUES
    ('organization_owner', 'Organization Owner'),
    ('administrator', 'Administrator'),
    ('recruiter', 'Recruiter'),
    ('hiring_manager', 'Hiring Manager'),
    ('interviewer', 'Interviewer'),
    ('approver', 'Approver'),
    ('candidate', 'Candidate');

-- Enable tenant-level row security.
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE member_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;

CREATE POLICY organizations_tenant_isolation
    ON organizations
    USING (
        id = NULLIF(
            current_setting('app.current_organization_id', TRUE), ''
        )::UUID
    )
    WITH CHECK (
        id = NULLIF(
            current_setting('app.current_organization_id', TRUE), ''
        )::UUID
    );

CREATE POLICY organization_members_tenant_isolation
    ON organization_members
    USING (
        organization_id = NULLIF(
            current_setting('app.current_organization_id', TRUE), ''
        )::UUID
    )
    WITH CHECK (
        organization_id = NULLIF(
            current_setting('app.current_organization_id', TRUE), ''
        )::UUID
    );

CREATE POLICY member_roles_tenant_isolation
    ON member_roles
    USING (
        organization_id = NULLIF(
            current_setting('app.current_organization_id', TRUE), ''
        )::UUID
    )
    WITH CHECK (
        organization_id = NULLIF(
            current_setting('app.current_organization_id', TRUE), ''
        )::UUID
    );

CREATE POLICY invitations_tenant_isolation
    ON invitations
    USING (
        organization_id = NULLIF(
            current_setting('app.current_organization_id', TRUE), ''
        )::UUID
    )
    WITH CHECK (
        organization_id = NULLIF(
            current_setting('app.current_organization_id', TRUE), ''
        )::UUID
    );

-- Candidate accounts are scoped to their authenticated user.
ALTER TABLE candidate_accounts ENABLE ROW LEVEL SECURITY;

CREATE POLICY candidate_accounts_user_isolation
    ON candidate_accounts
    USING (
        user_id = NULLIF(
            current_setting('app.current_user_id', TRUE), ''
        )::UUID
    )
    WITH CHECK (
        user_id = NULLIF(
            current_setting('app.current_user_id', TRUE), ''
        )::UUID
    );
