-- Included by permissions.sql inside its provisioning transaction. Serialize
-- session writes while installing the trigger and backfilling existing rows.
LOCK TABLE public.auth_session IN SHARE ROW EXCLUSIVE MODE;

-- Move the earlier schema's table rather than rebuilding it from active
-- sessions: history for sessions that have already ended must also survive.
DO $$
BEGIN
  IF to_regclass('keepfin_security.auth_login_history') IS NOT NULL THEN
    ALTER TABLE keepfin_security.auth_login_history SET SCHEMA public;
    ALTER TABLE public.auth_login_history OWNER TO auth_migrator;
  END IF;
END $$;

SET LOCAL ROLE auth_migrator;
CREATE TABLE IF NOT EXISTS public.auth_login_history (
  session_id text PRIMARY KEY,
  auth_user_id text NOT NULL,
  signed_in_at timestamptz NOT NULL,
  recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  ip_address text,
  user_agent text
);
CREATE INDEX IF NOT EXISTS auth_login_history_user_time_idx
  ON public.auth_login_history (auth_user_id, signed_in_at DESC, session_id);

-- No session/user foreign key: logout, expiry cleanup and identity deletion
-- must not cascade into this audit history. Never store session tokens here.
REVOKE ALL ON public.auth_login_history
  FROM PUBLIC, identity_operator, keepfin_runtime, auth_runtime, keepfin_migrator;

CREATE OR REPLACE FUNCTION public.record_auth_login()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER
SET search_path = pg_catalog, pg_temp
AS $$
BEGIN
  INSERT INTO public.auth_login_history
    (session_id, auth_user_id, signed_in_at, ip_address, user_agent)
  VALUES (NEW.id, NEW."userId", NEW."createdAt", NULLIF(NEW."ipAddress", ''), NULLIF(NEW."userAgent", ''));
  RETURN NEW;
END $$;
REVOKE ALL ON FUNCTION public.record_auth_login()
  FROM PUBLIC, identity_operator, keepfin_runtime, auth_runtime, keepfin_migrator;
RESET ROLE;

CREATE OR REPLACE TRIGGER auth_session_login_history
  AFTER INSERT ON public.auth_session
  FOR EACH ROW EXECUTE FUNCTION public.record_auth_login();
DROP FUNCTION IF EXISTS keepfin_security.record_auth_login();

-- Existing rows retain their original sign-in time; recorded_at marks when
-- history was first captured. Reapplying setup never overwrites a snapshot.
INSERT INTO public.auth_login_history
  (session_id, auth_user_id, signed_in_at, ip_address, user_agent)
SELECT id, "userId", "createdAt", NULLIF("ipAddress", ''), NULLIF("userAgent", '')
  FROM public.auth_session
ON CONFLICT (session_id) DO NOTHING;
