\set ON_ERROR_STOP on
BEGIN;
SELECT pg_advisory_xact_lock(hashtext('keepfin-security-bootstrap'));
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC, keepfin_runtime, auth_runtime, identity_operator;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM PUBLIC, keepfin_runtime, auth_runtime, identity_operator;
DO $$
DECLARE obj record; runtime_role text;
BEGIN
  FOR obj IN SELECT c.relname, c.relkind, r.rolname FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
             JOIN pg_roles r ON r.oid = c.relowner
             WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p', 'S') LOOP
    IF obj.rolname NOT IN ('auth_migrator', 'keepfin_migrator') THEN
      RAISE EXCEPTION 'Unexpected owner for public.%: run roles.sql first', obj.relname;
    END IF;
    runtime_role := CASE WHEN obj.rolname = 'auth_migrator' THEN 'auth_runtime' ELSE 'keepfin_runtime' END;
    IF obj.relkind = 'S' THEN
      EXECUTE format('GRANT USAGE, SELECT ON SEQUENCE public.%I TO %I', obj.relname, runtime_role);
    ELSE
      EXECUTE format('GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.%I TO %I', obj.relname, runtime_role);
    END IF;
  END LOOP;
END $$;
-- Existing identities and roles cannot be changed by a profile update.
REVOKE UPDATE, DELETE ON public.users FROM keepfin_runtime;
GRANT UPDATE (theme, base_currency, locale, updated_at, deleted_at) ON public.users TO keepfin_runtime;
GRANT SELECT (id, email, auth_user_id, deleted_at, user_status) ON public.users TO keepfin_security_owner;
GRANT UPDATE (auth_user_id, updated_at) ON public.users TO keepfin_security_owner;
GRANT SELECT (id, email, "emailVerified") ON public.auth_user TO keepfin_security_owner;
-- PostgreSQL requires an UPDATE privilege to take the verification row lock.
GRANT UPDATE (id) ON public.auth_user TO keepfin_security_owner;
\ir identity.sql
\ir login-history.sql
COMMIT;
