SET statement_timeout = 0;

--bun:split
CREATE TABLE ip_addresses(
	id SERIAL PRIMARY KEY,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	deleted_at TIMESTAMPTZ,
	ip_address TEXT NOT NULL,
	permission TEXT NOT NULL, -- 'allow_all' | 'priority' | 'deny' are the only possible values.
	comment TEXT
);

--bun:split
CREATE UNIQUE INDEX unique_ip_address ON ip_addresses(ip_address)
WHERE (deleted_at IS NULL);

