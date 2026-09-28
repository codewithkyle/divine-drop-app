package helpers

import (
    "github.com/gofiber/fiber/v2/log"
    "gorm.io/gorm"
)

// Exec runs a write and reports failures. Call sites used to discard the error
// entirely, so a failed INSERT or UPDATE was indistinguishable from a
// successful one.
func Exec(db *gorm.DB, query string, args ...interface{}) error {
    if err := db.Exec(query, args...).Error; err != nil {
        log.Error("Query failed", "error", err, "query", query)
        return err
    }
    return nil
}
