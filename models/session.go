package models

import (
    "gorm.io/gorm"
)

type Session struct {
    Id            string    `gorm:"column:session_id;primary_key;type:binary(16)"`
    UserId        string    `gorm:"column:user_id"`
    Data          []byte    `gorm:"column:data;type:blob"`
    Expires       []uint8   `gorm:"column:expires"`
}

// DeleteSession invalidates a session server side. Clearing the cookie alone
// leaves the row valid for the rest of its 168 hours.
func DeleteSession(db *gorm.DB, sessionId string) error {
    return db.Exec("DELETE FROM Sessions WHERE session_id = UNHEX(?)", sessionId).Error
}

// DeleteExpiredSessions drops rows past their expiry. Nothing else removes
// them, so the table only ever grew.
func DeleteExpiredSessions(db *gorm.DB) (int64, error) {
    res := db.Exec("DELETE FROM Sessions WHERE expires < NOW()")
    return res.RowsAffected, res.Error
}
