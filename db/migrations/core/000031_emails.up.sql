ALTER TABLE USERS ADD COLUMN email VARCHAR(255);
ALTER TABLE USERS ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE USERS ADD COLUMN email_unsubscriptions JSONB;
