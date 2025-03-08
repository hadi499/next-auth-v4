package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100)" json:"name"`
	Username  string    `gorm:"type:varchar(100);unique" json:"username"`
	Email     string    `gorm:"type:varchar(100);unique" json:"email"`
	Password  string    `gorm:"type:varchar(255)" json:"password,omitempty"`
	Role      string    `gorm:"type:varchar(50);default:'user'" json:"role"`
	Provider  string    `gorm:"type:varchar(50);default:'credentials'" json:"provider"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Fungsi BeforeCreate untuk menghasilkan UUID sebelum penyimpanan ke database
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.Id = uuid.New()
	return
}
