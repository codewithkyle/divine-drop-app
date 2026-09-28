package helpers

import (
	"database/sql"
	"errors"
	"os"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
	dbErr  error
)

// InitDB opens the one shared connection pool. Call it once, from main.
//
// Opening a pool per request leaks connections: *sql.DB has no finalizer that
// closes sockets, so every abandoned pool holds its connections open until
// MySQL's wait_timeout reaps them, and the server eventually hits
// max_connections.
func InitDB() error {
	dbOnce.Do(func() {
		db, dbErr = gorm.Open(mysql.Open(os.Getenv("DSN")), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if dbErr != nil {
			return
		}

		var pool *sql.DB
		pool, dbErr = db.DB()
		if dbErr != nil {
			return
		}
		pool.SetMaxOpenConns(25)
		pool.SetMaxIdleConns(10)
		pool.SetConnMaxLifetime(time.Hour)
		pool.SetConnMaxIdleTime(10 * time.Minute)
	})
	return dbErr
}

// ConnectDB returns the shared handle. *gorm.DB is safe for concurrent use, and
// every call site uses it as db.Raw(...) or db.Exec(...), both of which clone a
// fresh statement per call.
func ConnectDB() *gorm.DB {
	return db
}

// PingDB reports whether the shared pool can still reach the database. Used by
// the health endpoint so the proxy can tell "starting" from "gone".
func PingDB() error {
	if db == nil {
		return errors.New("database not initialised")
	}
	pool, err := db.DB()
	if err != nil {
		return err
	}
	return pool.Ping()
}
