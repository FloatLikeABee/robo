package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"idongivaflyinfa/config"
)

// CLIStores holds SQL + entity-detail handles for Morph cmd tools.
type CLIStores struct {
	SQL     *sql.DB
	Details EntityDetailStore
	Backend string // "embedded" or "legacy"
	closeFns []func()
}

// Close closes all opened stores.
func (s *CLIStores) Close() {
	if s == nil {
		return
	}
	for i := len(s.closeFns) - 1; i >= 0; i-- {
		if s.closeFns[i] != nil {
			s.closeFns[i]()
		}
	}
}

// OpenCLIStores opens Tran SQL and entity details for seed/backfill/prune CLIs.
// Default (STORAGE_BACKEND unset or "embedded"): SQLite + Badger.
// STORAGE_BACKEND=legacy: MySQL + Mongo (requires TRAN_MYSQL_DSN / TRAN_MONGO_*).
func OpenCLIStores(cfg config.Config) (*CLIStores, error) {
	backend := strings.ToLower(strings.TrimSpace(cfg.StorageBackend))
	if backend == "" {
		backend = "embedded"
	}
	out := &CLIStores{Backend: backend}

	if backend == "legacy" {
		dsn := strings.TrimSpace(cfg.TranMySQLDSN)
		if dsn == "" {
			return nil, fmt.Errorf("STORAGE_BACKEND=legacy requires TRAN_MYSQL_DSN")
		}
		m, err := NewTranMySQLLegacy(dsn)
		if err != nil {
			return nil, err
		}
		out.SQL = m.DB
		out.closeFns = append(out.closeFns, func() { _ = m.Close() })

		if strings.TrimSpace(cfg.TranMongoURI) != "" {
			mongo, err := NewTranMongo(cfg.TranMongoURI, cfg.TranMongoDB)
			if err != nil {
				out.Close()
				return nil, err
			}
			out.Details = mongo
			out.closeFns = append(out.closeFns, func() { _ = mongo.Close(context.Background()) })
		}
		return out, nil
	}

	m, err := NewTranSQL(cfg.TranSQLitePath)
	if err != nil {
		return nil, err
	}
	out.SQL = m.DB
	out.closeFns = append(out.closeFns, func() { _ = m.Close() })

	details, err := NewBadgerEntityDetails(cfg.EntityDetailsBadger)
	if err != nil {
		out.Close()
		return nil, err
	}
	out.Details = details
	out.closeFns = append(out.closeFns, func() { _ = details.Close() })
	return out, nil
}
