\set ON_ERROR_STOP on
\getenv keepfin_password KEEPFIN_DB_PASSWORD
\getenv keepfin_migration_password KEEPFIN_MIGRATION_DB_PASSWORD
\getenv auth_password AUTH_DB_PASSWORD
\getenv auth_migration_password AUTH_MIGRATION_DB_PASSWORD

BEGIN;
SELECT pg_advisory_xact_lock(hashtext('keepfin-security-bootstrap'));
DO $$
DECLARE role_name text;
BEGIN
  FOREACH role_name IN ARRAY ARRAY['keepfin_runtime', 'keepfin_migrator', 'auth_runtime', 'auth_migrator', 'identity_operator', 'keepfin_security_owner'] LOOP
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = role_name) THEN
      EXECUTE format('CREATE ROLE %I', role_name);
    END IF;
    IF EXISTS (SELECT FROM pg_auth_members m JOIN pg_roles r ON r.oid = m.member WHERE r.rolname = role_name) THEN
      RAISE EXCEPTION 'Managed role % must not inherit other roles', role_name;
    END IF;
    EXECUTE format('ALTER ROLE %I NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS', role_name);
  END LOOP;
END $$;
ALTER ROLE identity_operator NOLOGIN;
ALTER ROLE keepfin_security_owner NOLOGIN;
SELECT format('ALTER ROLE keepfin_runtime LOGIN PASSWORD %L', :'keepfin_password') \gexec
SELECT format('ALTER ROLE keepfin_migrator LOGIN PASSWORD %L', :'keepfin_migration_password') \gexec
SELECT format('ALTER ROLE auth_runtime LOGIN PASSWORD %L', :'auth_password') \gexec
SELECT format('ALTER ROLE auth_migrator LOGIN PASSWORD %L', :'auth_migration_password') \gexec

-- Run only against a dedicated Keepfin database, including existing volumes.
SELECT format('REVOKE ALL ON DATABASE %I FROM PUBLIC', current_database()) \gexec
SELECT format('GRANT CONNECT ON DATABASE %I TO keepfin_runtime, keepfin_migrator, auth_runtime, auth_migrator, identity_operator', current_database()) \gexec
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO keepfin_runtime, keepfin_migrator, auth_runtime, auth_migrator, keepfin_security_owner;
GRANT CREATE ON SCHEMA public TO keepfin_migrator, auth_migrator;
CREATE SCHEMA IF NOT EXISTS keepfin_security AUTHORIZATION keepfin_security_owner;
REVOKE ALL ON SCHEMA keepfin_security FROM PUBLIC;
GRANT USAGE ON SCHEMA keepfin_security TO identity_operator;

-- Transfer existing objects without changing rows, UUIDs or foreign keys.
DO $$
DECLARE obj record; owner_name text;
BEGIN
  FOR obj IN SELECT c.relname, c.relkind FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
             WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p', 'S') ORDER BY (c.relkind = 'S') LOOP
    owner_name := CASE WHEN obj.relname LIKE 'auth\_%' ESCAPE '\' THEN 'auth_migrator' ELSE 'keepfin_migrator' END;
    EXECUTE format('ALTER %s public.%I OWNER TO %I', CASE WHEN obj.relkind = 'S' THEN 'SEQUENCE' ELSE 'TABLE' END, obj.relname, owner_name);
  END LOOP;
  FOR obj IN SELECT p.oid::regprocedure AS signature FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace
             WHERE n.nspname = 'public' AND p.proname IN ('keepfin_holdings_search_dirty', 'keepfin_account_search_dirty', 'keepfin_notify_worker') LOOP
    EXECUTE format('ALTER FUNCTION %s OWNER TO keepfin_migrator', obj.signature);
  END LOOP;
END $$;

ALTER DEFAULT PRIVILEGES FOR ROLE keepfin_migrator IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO keepfin_runtime;
ALTER DEFAULT PRIVILEGES FOR ROLE keepfin_migrator IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO keepfin_runtime;
ALTER DEFAULT PRIVILEGES FOR ROLE auth_migrator IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO auth_runtime;
ALTER DEFAULT PRIVILEGES FOR ROLE auth_migrator IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO auth_runtime;
ALTER DEFAULT PRIVILEGES FOR ROLE keepfin_migrator REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE auth_migrator REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE keepfin_security_owner REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
COMMIT;
