package services

import (
	"context"
	"fmt"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
)

type paymentService struct {
	paymentRepo         output.PaymentRepository
	membershipFeeRepo   output.MembershipFeeRepository
	memberRepo          output.MemberRepository
	notificationService NotificationService
	feeCalculator       input.FeeCalculator
}

func NewPaymentService(
	paymentRepo output.PaymentRepository,
	membershipFeeRepo output.MembershipFeeRepository,
	memberRepo output.MemberRepository,
	notificationService NotificationService,
	feeCalculator input.FeeCalculator,
) input.PaymentService {
	return &paymentService{
		paymentRepo:         paymentRepo,
		membershipFeeRepo:   membershipFeeRepo,
		memberRepo:          memberRepo,
		notificationService: notificationService,
		feeCalculator:       feeCalculator,
	}
}

func (s *paymentService) RegisterPayment(ctx context.Context, payment *models.Payment) error {
	if err := payment.Validate(); err != nil {
		return &input.PaymentError{
			Code:    "INVALID_PAYMENT",
			Message: err.Error(),
		}
	}

	// Si es un pago de cuota, actualizar el estado de la cuota
	if payment.MembershipFeeID != nil {
		fee, err := s.membershipFeeRepo.FindByYearMonth(ctx, time.Now().Year(), int(time.Now().Month()))
		if err != nil {
			return fmt.Errorf("error finding membership fee: %w", err)
		}
		fee.Status = models.PaymentStatusPaid
		if err := s.membershipFeeRepo.Update(ctx, fee); err != nil {
			return fmt.Errorf("error updating membership fee: %w", err)
		}
	}

	return s.paymentRepo.Create(ctx, payment)
}

func (s *paymentService) CancelPayment(ctx context.Context, paymentID uint, reason string) error {
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("error finding payment: %w", err)
	}
	if payment == nil {
		return &input.PaymentError{
			Code:    "PAYMENT_NOT_FOUND",
			Message: "payment not found",
		}
	}

	payment.Status = models.PaymentStatusCancelled
	payment.Notes = reason
	return s.paymentRepo.Update(ctx, payment)
}

func (s *paymentService) GetPayment(ctx context.Context, paymentID uint) (*models.Payment, error) {
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("error finding payment: %w", err)
	}
	if payment == nil {
		return nil, &input.PaymentError{
			Code:    "PAYMENT_NOT_FOUND",
			Message: "payment not found",
		}
	}
	return payment, nil
}

func (s *paymentService) GenerateMonthlyFees(ctx context.Context, year, month int, baseAmount float64) error {
	// Generar cuota base
	fee := &models.MembershipFee{
		Year:           year,
		Month:          month,
		BaseFeeAmount:  baseAmount,
		FamilyFeeExtra: s.feeCalculator.CalculateFamilyFee(year, month) - baseAmount,
		Status:         models.PaymentStatusPending,
		DueDate:        time.Date(year, time.Month(month), 10, 0, 0, 0, 0, time.Local),
	}

	return s.membershipFeeRepo.Create(ctx, fee)
}

func (s *paymentService) GetMembershipFee(ctx context.Context, year, month int) (*models.MembershipFee, error) {
	fee, err := s.membershipFeeRepo.FindByYearMonth(ctx, year, month)
	if err != nil {
		return nil, fmt.Errorf("error finding membership fee: %w", err)
	}
	if fee == nil {
		return nil, &input.PaymentError{
			Code:    "FEE_NOT_FOUND",
			Message: "membership fee not found",
		}
	}
	return fee, nil
}

func (s *paymentService) UpdateFeeAmount(ctx context.Context, feeID uint, newAmount float64) error {
	fee, err := s.membershipFeeRepo.FindByYearMonth(ctx, time.Now().Year(), int(time.Now().Month()))
	if err != nil {
		return fmt.Errorf("error finding membership fee: %w", err)
	}
	if fee == nil {
		return &input.PaymentError{
			Code:    "FEE_NOT_FOUND",
			Message: "membership fee not found",
		}
	}

	fee.BaseFeeAmount = newAmount
	return s.membershipFeeRepo.Update(ctx, fee)
}

func (s *paymentService) GetMemberStatement(ctx context.Context, memberID uint) (*input.AccountStatement, error) {
	// Obtener pagos del último año
	from := time.Now().AddDate(-1, 0, 0)
	payments, err := s.paymentRepo.FindByMember(ctx, memberID, from, time.Now())
	if err != nil {
		return nil, fmt.Errorf("error finding member payments: %w", err)
	}

	// Obtener cuotas pendientes
	pendingFees, err := s.membershipFeeRepo.FindPendingByMember(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("error finding pending fees: %w", err)
	}

	// Calcular total pagado
	var totalPaid float64
	var lastPaymentDate *time.Time
	for _, p := range payments {
		totalPaid += p.Amount
		if lastPaymentDate == nil || p.PaymentDate.After(*lastPaymentDate) {
			lastPaymentDate = &p.PaymentDate
		}
	}

	// Determinar si es moroso
	isDefaulter := false
	defaultDays := 0
	if len(pendingFees) > 0 {
		oldestPending := pendingFees[0]
		if time.Now().After(oldestPending.DueDate) {
			isDefaulter = true
			defaultDays = int(time.Since(oldestPending.DueDate).Hours() / 24)
		}
	}

	return &input.AccountStatement{
		TotalPaid:       totalPaid,
		PendingPayments: pendingFees,
		PaymentHistory:  payments,
		LastPaymentDate: lastPaymentDate,
		IsDefaulter:     isDefaulter,
		DefaultDays:     defaultDays,
	}, nil
}

func (s *paymentService) GetFamilyStatement(ctx context.Context, familyID uint) (*input.AccountStatement, error) {
	// Similar a GetMemberStatement pero para familias
	from := time.Now().AddDate(-1, 0, 0)
	payments, err := s.paymentRepo.FindByFamily(ctx, familyID, from, time.Now())
	if err != nil {
		return nil, fmt.Errorf("error finding family payments: %w", err)
	}

	var totalPaid float64
	var lastPaymentDate *time.Time
	for _, p := range payments {
		totalPaid += p.Amount
		if lastPaymentDate == nil || p.PaymentDate.After(*lastPaymentDate) {
			lastPaymentDate = &p.PaymentDate
		}
	}

	return &input.AccountStatement{
		TotalPaid:       totalPaid,
		PaymentHistory:  payments,
		LastPaymentDate: lastPaymentDate,
	}, nil
}

func (s *paymentService) GetDefaulters(ctx context.Context) ([]input.AccountStatement, error) {
	var defaulters []input.AccountStatement

	// Obtener todas las cuotas vencidas
	now := time.Now()
	pendingFees, err := s.membershipFeeRepo.FindPendingByMember(ctx, 0) // 0 para obtener todos
	if err != nil {
		return nil, err
	}

	// Agrupar por miembro y calcular días de atraso
	memberMap := make(map[uint]bool)
	for _, fee := range pendingFees {
		if !fee.DueDate.Before(now) {
			continue
		}

		if !memberMap[fee.Payment.MemberID] {
			statement, err := s.GetMemberStatement(ctx, fee.Payment.MemberID)
			if err != nil {
				return nil, err
			}
			if statement.IsDefaulter {
				defaulters = append(defaulters, *statement)
				memberMap[fee.Payment.MemberID] = true
			}
		}
	}

	return defaulters, nil
}

func (s *paymentService) SendPaymentReminder(ctx context.Context, memberID uint) error {
	// Implementación pendiente del sistema de notificaciones
	return nil
}

func (s *paymentService) SendPaymentConfirmation(ctx context.Context, paymentID uint) error {
	// Implementación pendiente del sistema de notificaciones
	return nil
}

func (s *paymentService) SendDefaulterNotification(ctx context.Context, memberID uint, days int) error {
	// Implementación pendiente del sistema de notificaciones
	return nil
}
