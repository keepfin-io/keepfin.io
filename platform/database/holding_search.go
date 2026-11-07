package database

import "gorm.io/gorm"

func initializeHoldingSearch(db *gorm.DB) error {
	// Triggers cover direct writes, soft deletion, CSV imports and all ledger
	// mutations. A rollback also rolls back the outbox revision.
	return db.Exec(`
CREATE OR REPLACE FUNCTION keepfin_holdings_search_dirty() RETURNS trigger AS $$
DECLARE aid uuid;
BEGIN
  aid := CASE WHEN TG_OP = 'DELETE' THEN OLD.account_id ELSE NEW.account_id END;
  INSERT INTO holding_search_states (account_id, user_id, revision, indexed_revision, updated_at)
    SELECT id, user_id, 1, 0, now() FROM accounts WHERE id = aid
  ON CONFLICT (account_id) DO UPDATE SET
    revision = holding_search_states.revision + 1, user_id = EXCLUDED.user_id, updated_at = now();
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS keepfin_holdings_search_dirty ON holdings;
CREATE TRIGGER keepfin_holdings_search_dirty AFTER INSERT OR UPDATE OR DELETE ON holdings
FOR EACH ROW EXECUTE FUNCTION keepfin_holdings_search_dirty();

CREATE OR REPLACE FUNCTION keepfin_account_search_dirty() RETURNS trigger AS $$
BEGIN
  INSERT INTO holding_search_states (account_id, user_id, revision, indexed_revision, updated_at)
  VALUES (CASE WHEN TG_OP = 'DELETE' THEN OLD.id ELSE NEW.id END,
          CASE WHEN TG_OP = 'DELETE' THEN OLD.user_id ELSE NEW.user_id END, 1, 0, now())
  ON CONFLICT (account_id) DO UPDATE SET
    revision = holding_search_states.revision + 1, user_id = EXCLUDED.user_id, updated_at = now();
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS keepfin_account_search_dirty ON accounts;
CREATE TRIGGER keepfin_account_search_dirty AFTER INSERT OR UPDATE OF deleted_at, user_id OR DELETE ON accounts
FOR EACH ROW EXECUTE FUNCTION keepfin_account_search_dirty();

INSERT INTO holding_search_states (account_id, user_id, revision, indexed_revision, updated_at)
SELECT id, user_id, 1, 0, now() FROM accounts
ON CONFLICT (account_id) DO NOTHING;
`).Error
}
