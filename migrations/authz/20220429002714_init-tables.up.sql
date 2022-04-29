CREATE TABLE authz_addresses (
    id SERIAL PRIMARY KEY, -- Maybe I should just do UUID PKs
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    wallet_address TEXT NOT NULL,
    permission TEXT NOT NULL -- 'allow' | 'deny' are the only possible values. 
);

--bun:split

CREATE UNIQUE INDEX unique_wallet_address ON authz_addresses (wallet_address) WHERE (deleted_at is NOT null);

--bun:split

CREATE INDEX CONCURRENTLY authz_addresses_deleted_at ON public.authz_addresses (deleted_at);

--bun:split

CREATE INDEX CONCURRENTLY authz_addresses_permission ON public.authz_addresses (permission);
