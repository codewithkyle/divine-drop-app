package models

type Session struct {
    Id            string    `gorm:"column:session_id;primary_key;type:binary(16)"`
    UserId        string    `gorm:"column:user_id"`
    Data          []byte    `gorm:"column:data;type:blob"`
    Expires       []uint8   `gorm:"column:expires"`
}
