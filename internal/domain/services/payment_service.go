package services

import (
	"context"
	"strconv"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/metrics"
)

type paymentService struct {
	paymentRepo         output.PaymentRepository
	membershipFeeRepo   output.MembershipFeeRepository
	memberRepo          output.MemberRepository
	notificationService input.NotificationService
	feeCalculator       input.FeeCalculator
}

func NewPaymentService(
	paymentRepo output.PaymentRepository,
	membershipFeeRepo output.MembershipFeeRepository,
	memberRepo output.MemberRepository,
	notificationService input.NotificationService,
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
	// Validar el pago
	if err := payment.Validate(); err != nil {
		// Si ya es un AppError, retornarlo directamente
		if appErr, ok := errors.AsAppError(err); ok {
			return appErr
		}
		// Sino, convertirlo a un error de validación estructurado
		return errors.NewValidationError(err.Error(), nil)
	}

	// Obtener miembro para validar que existe y está activo
	member, err := s.memberRepo.GetByID(ctx, payment.MemberID)
	if err != nil {
		return errors.DB(err, "error obteniendo miembro")
	}

	if member == nil {
		return errors.NotFound("member", nil)
	}

	// Si es un pago de cuota, actualizar el estado de la cuota
	if payment.MembershipFeeID != nil {
		// Buscar cuota existente
		fee, err := s.membershipFeeRepo.FindByYearMonth(ctx, time.Now().Year(), int(time.Now().Month()))
		if err != nil {
			return errors.DB(err, "error buscando cuota de membresía")
		}

		if fee == nil {
			return errors.NotFound("membership fee", nil)
		}

		// Actualizar estado de la cuota
		fee.Status = models.PaymentStatusPaid
		if err := s.membershipFeeRepo.Update(ctx, fee); err != nil {
			return errors.DB(err, "error actualizando cuota de membresía")
		}
	}

	// Crear el payment
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return errors.DB(err, "error creando pago")
	}

	// Registrar métricas del pago
	metrics.PaymentMetrics.WithLabelValues(
		paymentTypeToString(payment.MembershipFeeID != nil),
		string(payment.Status),
	).Set(payment.Amount)

	// Si es un pago atrasado, registrar la latencia
	if payment.PaymentDate.Before(payment.PaymentDate) {
		daysLate := payment.PaymentDate.Sub(payment.PaymentDate).Hours() / 24
		metrics.PaymentLatency.WithLabelValues(
			s.memberTypeToString(member),
		).Observe(daysLate)
	}

	return nil
}

// Funciones auxiliares para las métricas
func paymentTypeToString(isMembershipFee bool) string {
	if isMembershipFee {
		return "cuota"
	}
	return "otros"
}

// Función auxiliar para obtener el tipo de membresía
// Versión optimizada que evita una segunda llamada a la base de datos
func (s *paymentService) memberTypeToString(member *models.Member) string {
	if member == nil {
		// Si no encontramos el miembro, retornamos un valor por defecto
		return models.TipoMembresiaPIndividual
	}
	return member.TipoMembresia
}

func (s *paymentService) CancelPayment(ctx context.Context, paymentID uint, reason string) error {
	// Buscar el pago existente
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return errors.DB(err, "error buscando pago")
	}

	if payment == nil {
		return errors.NotFound("payment", nil)
	}

	// Actualizar el estado del pago
	payment.Status = models.PaymentStatusCancelled
	payment.Notes = reason

	return s.paymentRepo.Update(ctx, payment)
}

func (s *paymentService) GetPayment(ctx context.Context, paymentID uint) (*models.Payment, error) {
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, errors.DB(err, "error buscando pago")
	}

	if payment == nil {
		return nil, errors.NotFound("payment", nil)
	}

	return payment, nil
}

func (s *paymentService) GetMemberPayments(ctx context.Context, memberID uint) ([]*models.Payment, error) {
	// Definir un rango de fechas amplio para búsqueda
	from := time.Time{} // Tiempo cero para "desde siempre"
	to := time.Now()

	// Verificar que el miembro existe
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, errors.DB(err, "error verificando miembro")
	}

	if member == nil {
		return nil, errors.NotFound("member", nil)
	}

	// Buscar pagos por memberID
	payments, err := s.paymentRepo.FindByMember(ctx, memberID, from, to)
	if err != nil {
		return nil, errors.DB(err, "error buscando pagos del miembro")
	}

	// Convertir []models.Payment a []*models.Payment
	paymentPtrs := make([]*models.Payment, len(payments))
	for i := range payments {
		paymentPtrs[i] = &payments[i]
	}

	return paymentPtrs, nil
}

func (s *paymentService) GetFamilyPayments(ctx context.Context, familyID uint) ([]*models.Payment, error) {
	// Definir un rango de fechas amplio para búsqueda
	from := time.Time{} // Tiempo cero para "desde siempre"
	to := time.Now()

	// Obtener pagos por familyID
	payments, err := s.paymentRepo.FindByFamily(ctx, familyID, from, to)
	if err != nil {
		return nil, errors.DB(err, "error buscando pagos de la familia")
	}

	// Convertir []models.Payment a []*models.Payment
	paymentPtrs := make([]*models.Payment, len(payments))
	for i := range payments {
		paymentPtrs[i] = &payments[i]
	}

	return paymentPtrs, nil
}

func (s *paymentService) GenerateMonthlyFees(ctx context.Context, year, month int, baseAmount float64) error {
	// Validar datos de entrada
	if baseAmount <= 0 {
		return errors.Validation("El monto base debe ser positivo", "baseAmount", "debe ser positivo")
	}

	// Verificar si ya existe una cuota para el mismo año/mes
	existingFee, err := s.membershipFeeRepo.FindByYearMonth(ctx, year, month)
	if err != nil {
		return errors.DB(err, "error verificando cuota existente")
	}

	if existingFee != nil {
		return errors.New(errors.ErrDuplicateEntry, "ya existe una cuota para este período")
	}

	// Generar cuota base
	fee := &models.MembershipFee{
		Year:           year,
		Month:          month,
		BaseFeeAmount:  baseAmount,
		FamilyFeeExtra: s.feeCalculator.CalculateFamilyFee(year, month) - baseAmount,
		Status:         models.PaymentStatusPending,
		DueDate:        time.Date(year, time.Month(month), 10, 0, 0, 0, 0, time.UTC),
	}

	return s.membershipFeeRepo.Create(ctx, fee)
}

func (s *paymentService) GetMembershipFee(ctx context.Context, year, month int) (*models.MembershipFee, error) {
	fee, err := s.membershipFeeRepo.FindByYearMonth(ctx, year, month)
	if err != nil {
		return nil, errors.DB(err, "error buscando cuota de membresía")
	}

	if fee == nil {
		return nil, errors.NotFound("membership fee", nil)
	}

	return fee, nil
}

func (s *paymentService) UpdateFeeAmount(ctx context.Context, feeID uint, newAmount float64) error {
	// Validar monto
	if newAmount <= 0 {
		return errors.Validation("El monto debe ser positivo", "amount", "debe ser positivo")
	}

	// Obtener la cuota específica por ID
	fee, err := s.membershipFeeRepo.FindByID(ctx, feeID)
	if err != nil {
		return errors.DB(err, "error buscando cuota de membresía")
	}

	if fee == nil {
		return errors.NotFound("membership fee", nil)
	}

	// Actualizar monto de la cuota
	fee.BaseFeeAmount = newAmount
	return s.membershipFeeRepo.Update(ctx, fee)
}

func (s *paymentService) GetMemberStatement(ctx context.Context, memberID uint) (*input.AccountStatement, error) {
	// Verificar que el miembro existe
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, errors.DB(err, "error verificando miembro")
	}

	if member == nil {
		return nil, errors.NotFound("member", nil)
	}

	// Obtener pagos del último año
	from := time.Now().AddDate(-1, 0, 0)
	payments, err := s.paymentRepo.FindByMember(ctx, memberID, from, time.Now())
	if err != nil {
		return nil, errors.DB(err, "error buscando pagos del miembro")
	}

	// Obtener cuotas pendientes
	pendingFees, err := s.membershipFeeRepo.FindPendingByMember(ctx, memberID)
	if err != nil {
		return nil, errors.DB(err, "error buscando cuotas pendientes")
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
		return nil, errors.DB(err, "error buscando pagos de la familia")
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
		return nil, errors.DB(err, "error buscando cuotas pendientes")
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
				// Loguear el error pero continuar con otros miembros
				continue
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
	// Verificar que el miembro existe
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return errors.DB(err, "error obteniendo miembro")
	}

	if member == nil {
		return errors.NotFound("member", nil)
	}

	// Enviar notificación usando el servicio de notificaciones
	if member.CorreoElectronico != nil {
		return s.notificationService.SendEmail(
			ctx,
			*member.CorreoElectronico,
			"Recordatorio de Pago ASAM",
			"Este es un recordatorio para realizar su pago pendiente.",
		)
	}
	return nil
}

func (s *paymentService) SendPaymentConfirmation(ctx context.Context, paymentID uint) error {
	// Obtener información del pago y miembro
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return errors.DB(err, "error obteniendo pago")
	}

	if payment == nil {
		return errors.NotFound("payment", nil)
	}

	member, err := s.memberRepo.GetByID(ctx, payment.MemberID)
	if err != nil {
		return errors.DB(err, "error obteniendo miembro")
	}

	if member == nil {
		return errors.NotFound("member", nil)
	}

	// Enviar confirmación por email
	if member.CorreoElectronico != nil {
		amountStr := strconv.FormatFloat(payment.Amount, 'f', 2, 64)
		return s.notificationService.SendEmail(
			ctx,
			*member.CorreoElectronico,
			"Confirmación de Pago ASAM",
			"Se ha registrado su pago por "+amountStr+"€",
		)
	}
	return nil
}

func (s *paymentService) SendDefaulterNotification(ctx context.Context, memberID uint, days int) error {
	// Verificar que el miembro existe
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return errors.DB(err, "error obteniendo miembro")
	}

	if member == nil {
		return errors.NotFound("member", nil)
	}

	// Enviar notificación con días de retraso
	if member.CorreoElectronico != nil {
		daysStr := strconv.Itoa(days)
		return s.notificationService.SendEmail(
			ctx,
			*member.CorreoElectronico,
			"Aviso de Pago Atrasado ASAM",
			"Su pago está atrasado "+daysStr+" días. Por favor, regularice su situación.",
		)
	}
	return nil
}
