package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

// Role define los tipos de roles disponibles
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// User representa un usuario del sistema con autenticación
type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null"`
	Password     string `gorm:"not null"`
	Role         Role   `gorm:"type:varchar(20);not null;default:'user'"`
	LastLogin    time.Time
	IsActive     bool   `gorm:"not null;default:true"`
	RefreshToken string `gorm:"size:255"`
}

// SetPassword hashea y guarda la contraseña del usuario
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifica si la contraseña proporcionada es correcta
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsAdmin verifica si el usuario tiene rol de administrador
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// BeforeCreate hook de GORM que se ejecuta antes de crear un usuario
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.Role == "" {
		u.Role = RoleUser
	}
	return nil
}
