package database

import "gorm.io/gorm"

func initializeWorkerNotifications(db *gorm.DB) error {
	// NOTIFY is delivered only after commit. Keep financial/identity data out of
	// its payload, which is visible to every listener in the same database.
	return db.Exec(`
CREATE OR REPLACE FUNCTION keepfin_notify_worker() RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify(TG_ARGV[0], TG_TABLE_SCHEMA);
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS keepfin_holdings_changed ON holding_search_states;
CREATE TRIGGER keepfin_holdings_changed AFTER INSERT OR UPDATE OF revision, indexed_revision ON holding_search_states
FOR EACH ROW WHEN (NEW.revision > NEW.indexed_revision)
EXECUTE FUNCTION keepfin_notify_worker('keepfin_holdings_changed');

DROP TRIGGER IF EXISTS keepfin_market_jobs_changed ON market_jobs;
CREATE TRIGGER keepfin_market_jobs_changed AFTER INSERT OR UPDATE OF status, available_at ON market_jobs
FOR EACH ROW WHEN (NEW.status = 'queued')
EXECUTE FUNCTION keepfin_notify_worker('keepfin_market_jobs_changed');

DROP TRIGGER IF EXISTS keepfin_market_config_changed ON market_configs;
CREATE TRIGGER keepfin_market_config_changed AFTER INSERT OR UPDATE OR DELETE ON market_configs
FOR EACH STATEMENT EXECUTE FUNCTION keepfin_notify_worker('keepfin_market_config_changed');
`).Error
}
