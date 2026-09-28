package helpers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	appdb "app/db"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"

	// Registers the "mysql" dbmate driver as a side effect.
	_ "github.com/amacneil/dbmate/v2/pkg/driver/mysql"

	mysqldsn "github.com/go-sql-driver/mysql"
)

const (
	// migrationLock serialises migrations across processes. dbmate has no
	// advisory locking on MySQL, so two servers booting together would
	// otherwise run the same pending migration twice.
	migrationLock = "divinedrop_migrations"
	// migrationLockWait is how long to wait for another booting process to
	// finish before giving up. Long enough for an ALTER on the larger tables.
	migrationLockWait = 5 * time.Minute
)

// RunMigrations applies every pending migration. Call it from main after
// InitDB and before the server listens, so a deploy carries its own schema
// changes to each environment instead of needing them applied by hand.
//
// Returning an error here is meant to stop the boot: serving requests against
// a schema the build does not expect fails in stranger ways than not starting.
func RunMigrations() error {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		return errors.New("DSN is not set")
	}

	databaseURL, err := dbmateURL(dsn)
	if err != nil {
		return err
	}

	mate := dbmate.New(databaseURL)
	mate.FS = appdb.Migrations
	mate.MigrationsDir = []string{"migrations"}
	// dbmate otherwise shells out to mysqldump to refresh db/schema.sql after
	// migrating. The runtime image has no mysqldump and no source tree to
	// write to.
	mate.AutoDumpSchema = false
	// Refuse a migration numbered below one already applied, which is what a
	// branch merged out of order looks like.
	mate.Strict = true

	return withMigrationLock(func() error {
		if err := mate.Migrate(); err != nil && !errors.Is(err, dbmate.ErrNoMigrationFiles) {
			return err
		}
		return nil
	})
}

// withMigrationLock runs fn while holding a named MySQL lock.
//
// GET_LOCK is scoped to a single session, so this pins one connection out of
// the pool for the duration instead of using the shared handle, which would
// hand the release to whichever connection happened to be free.
func withMigrationLock(fn func() error) error {
	if db == nil {
		return errors.New("database not initialised")
	}
	pool, err := db.DB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), migrationLockWait+time.Minute)
	defer cancel()

	conn, err := pool.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var acquired *int64
	row := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", migrationLock, int(migrationLockWait.Seconds()))
	if err := row.Scan(&acquired); err != nil {
		return fmt.Errorf("could not take the migration lock: %w", err)
	}
	// NULL means the attempt errored, 0 means it timed out waiting on another
	// process. Neither is safe to migrate through.
	if acquired == nil || *acquired != 1 {
		return errors.New("timed out waiting for another process to finish migrating")
	}

	defer func() {
		var released *int64
		_ = conn.QueryRowContext(context.Background(), "SELECT RELEASE_LOCK(?)", migrationLock).Scan(&released)
	}()

	return fn()
}

// dbmateURL converts the go-sql-driver DSN the app already runs on into the URL
// dbmate expects, so migrations use the configured database without a second
// connection string to keep in step in every environment.
func dbmateURL(dsn string) (*url.URL, error) {
	cfg, err := mysqldsn.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("could not parse DSN: %w", err)
	}

	databaseURL := &url.URL{
		Scheme: "mysql",
		Path:   "/" + cfg.DBName,
	}
	if cfg.User != "" {
		databaseURL.User = url.UserPassword(cfg.User, cfg.Passwd)
	}

	query := url.Values{}
	for key, value := range cfg.Params {
		query.Set(key, value)
	}
	if cfg.Net == "unix" {
		query.Set("socket", cfg.Addr)
	} else {
		databaseURL.Host = cfg.Addr
	}
	databaseURL.RawQuery = query.Encode()

	return databaseURL, nil
}
