package db

import (
	"context"
	"errors"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) output.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	if err := payment.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) Update(ctx context.Context, payment *models.Payment) error {
	if err := payment.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(payment).Error
}

func (r *paymentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Payment{}, id).Error
}

func (r *paymentRepository) FindByID(ctx context.Context, id uint) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.WithContext(ctx).First(&payment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindByMember(ctx context.Context, memberID uint, from, to time.Time) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND payment_date BETWEEN ? AND ?", memberID, from, to).
		Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) FindByFamily(ctx context.Context, familyID uint, from, to time.Time) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND payment_date BETWEEN ? AND ?", familyID, from, to).
		Find(&payments).Error
	return payments, err
}

type membershipFeeRepository struct {
	db *gorm.DB
}

func NewMembershipFeeRepository(db *gorm.DB) output.MembershipFeeRepository {
	return &membershipFeeRepository{db: db}
}

func (r *membershipFeeRepository) Create(ctx context.Context, fee *models.MembershipFee) error {
	return r.db.WithContext(ctx).Create(fee).Error
}

func (r *membershipFeeRepository) Update(ctx context.Context, fee *models.MembershipFee) error {
	return r.db.WithContext(ctx).Save(fee).Error
}

func (r *membershipFeeRepository) FindByYearMonth(ctx context.Context, year, month int) (*models.MembershipFee, error) {
	var fee models.MembershipFee
	if err := r.db.WithContext(ctx).
		Where("year = ? AND month = ?", year, month).
		First(&fee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &fee, nil
}

func (r *membershipFeeRepository) FindPendingByMember(ctx context.Context, memberID uint) ([]models.MembershipFee, error) {
	var fees []models.MembershipFee
	err := r.db.WithContext(ctx).
		Joins("LEFT JOIN payments ON membership_fees.payment_id = payments.id").
		Where("payments.member_id = ? AND membership_fees.status = ?",
			memberID, models.PaymentStatusPending).
		Find(&fees).Error
	return fees, err
}
