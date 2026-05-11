package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"siakad/backend/internal/config"
)

func OpenMySQL(ctx context.Context, cfg config.MySQLConfig) (*sql.DB, error) {
	driverCfg := mysqlDriver.NewConfig()
	driverCfg.Net = "tcp"
	driverCfg.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	driverCfg.User = cfg.User
	driverCfg.Passwd = cfg.Password
	driverCfg.DBName = cfg.Database
	driverCfg.ParseTime = cfg.ParseTime
	driverCfg.Collation = cfg.Collation
	driverCfg.Params = map[string]string{
		"charset": cfg.Charset,
	}

	if cfg.Loc != "" {
		driverCfg.Loc = mustLoadLocation(cfg.Loc)
	}

	db, err := sql.Open("mysql", driverCfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

func HealthCheck(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.PingContext(checkCtx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	return nil
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err == nil {
		return loc
	}

	return time.Local
}
