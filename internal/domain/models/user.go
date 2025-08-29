package models

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Role define los tipos de roles disponibles
type Role string

const (
	// RoleAdmin representa el rol de administrador
	RoleAdmin Role = "admin"
	// RoleUser representa el rol de usuario normal
	RoleUser Role = "user"
)

// User representa un usuario del sistema con autenticación
type User struct {
	gorm.Model
	Username                string  `gorm:"uniqueIndex:uni_users_username;not null;size:100"`
	Email                   string  `gorm:"uniqueIndex:uni_users_email;not null;size:255"`
	Password                string  `gorm:"not null"`
	Role                    Role    `gorm:"type:varchar(20);not null;default:'user'"`
	MemberID                *uint   `gorm:"index:,unique,where:member_id IS NOT NULL"`
	Member                  *Member `gorm:"foreignKey:MemberID;constraint:OnDelete:RESTRICT"`
	LastLogin               time.Time
	IsActive                bool `gorm:"not null;default:true"`
	EmailVerified           bool `gorm:"not null;default:false"`
	EmailVerifiedAt         *time.Time
	EmailVerificationSentAt *time.Time
	RefreshTokens           []RefreshToken      `gorm:"foreignKey:UserID"` // Relación con tokens
	VerificationTokens      []VerificationToken `gorm:"foreignKey:UserID"` // Relación con tokens de verificación
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
	// Trim any whitespace that might have been added
	trimmedPassword := strings.TrimSpace(password)
	trimmedHash := strings.TrimSpace(u.Password)

	err := bcrypt.CompareHashAndPassword([]byte(trimmedHash), []byte(trimmedPassword))
	return err == nil
}

// IsAdmin verifica si el usuario tiene rol de administrador
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// HasMemberAccess verifica si el usuario puede acceder a los datos de un socio específico
func (u *User) HasMemberAccess(memberID uint) bool {
	// Los administradores tienen acceso a todos los socios
	if u.IsAdmin() {
		return true
	}

	// Los usuarios regulares solo pueden acceder a su propio socio
	return u.MemberID != nil && *u.MemberID == memberID
}

// ValidateMemberAssociation valida la coherencia entre rol y asociación con socio
func (u *User) ValidateMemberAssociation() error {
	// Usuario con rol USER debe tener un socio asociado
	if u.Role == RoleUser && u.MemberID == nil {
		return errors.New("usuarios con rol USER deben tener un socio asociado")
	}

	// Usuario con rol ADMIN no puede tener socio asociado
	if u.Role == RoleAdmin && u.MemberID != nil {
		return errors.New("usuarios administradores no pueden tener socio asociado")
	}

	return nil
}

// BeforeCreate hook de GORM que se ejecuta antes de crear un usuario
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.Role == "" {
		u.Role = RoleUser
	}

	// Validar la asociación rol-socio
	return u.ValidateMemberAssociation()
}

// BeforeUpdate hook de GORM que se ejecuta antes de actualizar un usuario
func (u *User) BeforeUpdate(_ *gorm.DB) error {
	// Validar la asociación rol-socio
	return u.ValidateMemberAssociation()
}
