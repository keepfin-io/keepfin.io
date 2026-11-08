SET LOCAL ROLE keepfin_security_owner;
CREATE TABLE IF NOT EXISTS keepfin_security.identity_link_audit (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  occurred_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  operator_role text NOT NULL,
  client_address inet,
  application_user_id uuid NOT NULL,
  previous_auth_user_id text,
  auth_user_id text NOT NULL,
  reason text NOT NULL,
  ownership_reference text NOT NULL
);
REVOKE ALL ON keepfin_security.identity_link_audit FROM PUBLIC, identity_operator, keepfin_runtime, auth_runtime;

CREATE OR REPLACE FUNCTION keepfin_security.link_identity(
  application_id uuid, identity_id uuid, reason text, ownership_reference text
) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER
SET search_path = pg_catalog, keepfin_security, pg_temp
AS $$
DECLARE identity_record record; application_record record; audit_id uuid;
BEGIN
  IF NOT pg_has_role(session_user, 'identity_operator', 'MEMBER') THEN
    RAISE EXCEPTION 'Identity operator role required' USING ERRCODE = '42501';
  END IF;
  IF reason IS NULL OR length(btrim(reason)) NOT BETWEEN 10 AND 1000 OR
     ownership_reference IS NULL OR length(btrim(ownership_reference)) NOT BETWEEN 3 AND 200 THEN
    RAISE EXCEPTION 'A reason and verified ownership reference are required' USING ERRCODE = '22023';
  END IF;
  SELECT id, email, "emailVerified" INTO identity_record
    FROM public.auth_user WHERE id::text = identity_id::text FOR UPDATE;
  IF NOT FOUND OR NOT coalesce(identity_record."emailVerified", false) THEN
    RAISE EXCEPTION 'Identity mapping rejected: verified identity required';
  END IF;
  SELECT id, email, auth_user_id, deleted_at, user_status INTO application_record
    FROM public.users WHERE id = application_id FOR UPDATE;
  IF NOT FOUND OR application_record.deleted_at IS NOT NULL OR
     application_record.user_status IS DISTINCT FROM 1 OR
     lower(application_record.email) IS DISTINCT FROM lower(identity_record.email) OR
     (application_record.auth_user_id IS NOT NULL AND application_record.auth_user_id <> identity_id::text) THEN
    RAISE EXCEPTION 'Identity mapping rejected: check email, status and existing mapping';
  END IF;
  UPDATE public.users SET auth_user_id = identity_id::text, updated_at = clock_timestamp() WHERE id = application_id;
  INSERT INTO keepfin_security.identity_link_audit
    (operator_role, client_address, application_user_id, previous_auth_user_id, auth_user_id, reason, ownership_reference)
  VALUES (session_user, inet_client_addr(), application_id, application_record.auth_user_id,
          identity_id::text, btrim(reason), btrim(ownership_reference)) RETURNING id INTO audit_id;
  RETURN audit_id;
END $$;
REVOKE ALL ON FUNCTION keepfin_security.link_identity(uuid, uuid, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION keepfin_security.link_identity(uuid, uuid, text, text) TO identity_operator;
RESET ROLE;
