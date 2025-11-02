package services

import (
	"context"
	"log"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/metrics"
)

type paymentService struct {
	paymentRepo       output.PaymentRepository
	membershipFeeRepo output.MembershipFeeRepository
	memberRepo        output.MemberRepository
	familyRepo        output.FamilyRepository
	cashFlowRepo      output.CashFlowRepository
	feeCalculator     input.FeeCalculator
}

// NewPaymentService crea una nueva instancia del servicio de pagos
// que implementa la interfaz input.PaymentService.
func NewPaymentService(
	paymentRepo output.PaymentRepository,
	membershipFeeRepo output.MembershipFeeRepository,
	memberRepo output.MemberRepository,
	familyRepo output.FamilyRepository,
	cashFlowRepo output.CashFlowRepository,
	feeCalculator input.FeeCalculator,
) input.PaymentService {
	return &paymentService{
		paymentRepo:       paymentRepo,
		membershipFeeRepo: membershipFeeRepo,
		memberRepo:        memberRepo,
		familyRepo:        familyRepo,
		cashFlowRepo:      cashFlowRepo,
		feeCalculator:     feeCalculator,
	}
}

func (s *paymentService) RegisterPayment(ctx context.Context, payment *models.Payment) error {
	// Validar el pago
	if err := s.validatePayment(ctx, payment); err != nil {
		return err
	}

	// Si es un pago inicial (no tiene MembershipFeeID), asociar con cuota anual
	if payment.MembershipFeeID == nil {
		// Verificar que no exista ya un pago inicial para este miembro
		hasInitial, err := s.paymentRepo.HasInitialPayment(ctx, &payment.MemberID, nil)
		if err != nil {
			return errors.DB(err, "error verificando pagos existentes")
		}

		if hasInitial {
			return errors.NewValidationError(
				"Ya existe un pago inicial registrado para este socio",
				map[string]string{
					"duplicate": "initial_payment_already_exists",
				},
			)
		}

		if err := s.ensureAnnualFee(ctx, payment); err != nil {
			return err
		}
	}

	// Procesar pago de cuota si aplica
	if err := s.processMembershipFee(ctx, payment); err != nil {
		return err
	}

	// Crear el payment
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return errors.DB(err, "error creando pago")
	}

	// Registrar métricas del pago
	if err := s.recordPaymentMetrics(ctx, payment); err != nil {
		// Registrar el error sin interrumpir el flujo principal
		log.Printf("Error registrando métricas de pago: %v", err)
	}

	return nil
}

// ensureAnnualFee busca o crea la cuota anual del año actual y la asocia al pago
func (s *paymentService) ensureAnnualFee(ctx context.Context, payment *models.Payment) error {
	currentYear := time.Now().Year()

	// Buscar cuota anual existente
	fee, err := s.membershipFeeRepo.FindByYear(ctx, currentYear)
	if err != nil {
		return errors.DB(err, "error buscando cuota anual")
	}

	// Si no existe, crearla
	if fee == nil {
		// Usar el monto del pago como monto base de la cuota
		baseAmount := payment.Amount

		fee = models.NewAnnualFee(currentYear, baseAmount)

		if err := s.membershipFeeRepo.Create(ctx, fee); err != nil {
			return errors.DB(err, "error creando cuota anual")
		}
	}

	// Asociar el pago con la cuota anual
	payment.MembershipFeeID = &fee.ID
	payment.MembershipFee = fee

	return nil
}

// validatePayment valida que el pago sea correcto y que el miembro o familia exista
func (s *paymentService) validatePayment(ctx context.Context, payment *models.Payment) error {
	// Validar el pago
	if err := payment.Validate(); err != nil {
		// Si ya es un AppError, retornarlo directamente
		if appErr, ok := errors.AsAppError(err); ok {
			return appErr
		}
		// Sino, convertirlo a un error de validación estructurado
		return errors.NewValidationError(err.Error(), nil)
	}

	// Validar miembro (siempre requerido)
	return s.validateMember(ctx, payment.MemberID)
}

// validateMember verifica que el miembro exista y esté activo
func (s *paymentService) validateMember(ctx context.Context, memberID uint) error {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return errors.DB(err, "error obteniendo miembro")
	}

	if member == nil {
		return errors.NotFound("member", nil)
	}

	// Verificar que el miembro esté activo
	if member.State != models.EstadoActivo {
		return errors.Validation("El miembro no está activo", "estado", "inactive")
	}

	return nil
}

// processMembershipFee valida que la cuota de membresía exista
// Nota: El estado del pago se maneja en la tabla payments, no en membership_fees
func (s *paymentService) processMembershipFee(ctx context.Context, payment *models.Payment) error {
	// Si no es un pago de cuota, no hay nada que procesar
	if payment.MembershipFeeID == nil {
		return nil
	}

	// Buscar cuota existente para validar que existe
	fee, err := s.membershipFeeRepo.FindByID(ctx, *payment.MembershipFeeID)
	if err != nil {
		return errors.DB(err, "error buscando cuota de membresía")
	}

	if fee == nil {
		return errors.NotFound("membership fee", nil)
	}

	// No necesitamos actualizar nada en membership_fee
	// El estado del pago individual ya está en payment.Status
	return nil
}

// recordPaymentMetrics registra métricas relacionadas con el pago
func (s *paymentService) recordPaymentMetrics(ctx context.Context, payment *models.Payment) error {
	// Registrar métricas del pago
	metrics.PaymentMetrics.WithLabelValues(
		paymentTypeToString(payment.MembershipFeeID != nil),
		string(payment.Status),
	).Set(payment.Amount)

	// Verificar si el pago de cuota está atrasado
	if payment.MembershipFeeID != nil {
		return s.recordLatencyMetrics(ctx, payment)
	}

	return nil
}

// recordLatencyMetrics registra métricas de latencia para pagos de cuotas
func (s *paymentService) recordLatencyMetrics(ctx context.Context, payment *models.Payment) error {
	fee, err := s.membershipFeeRepo.FindByID(ctx, *payment.MembershipFeeID)
	if err != nil {
		return err // Error silencioso para métricas, no afecta el flujo principal
	}

	if fee != nil && payment.PaymentDate != nil && fee.DueDate.Before(*payment.PaymentDate) {
		// Obtener el miembro para las métricas
		member, err := s.memberRepo.GetByID(ctx, payment.MemberID)
		if err != nil {
			return err // Error silencioso para métricas
		}

		daysLate := payment.PaymentDate.Sub(fee.DueDate).Hours() / 24
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
	return member.MembershipType
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

	// Verificar si ya está cancelado
	if payment.Status == models.PaymentStatusCancelled {
		return errors.Validation("El pago ya está cancelado", "status", "already cancelled")
	}

	// Actualizar el estado del pago
	payment.Status = models.PaymentStatusCancelled
	payment.Notes = reason

	return s.paymentRepo.Update(ctx, payment)
}

// ConfirmPayment confirms a pending payment by changing its status to PAID
func (s *paymentService) ConfirmPayment(ctx context.Context, paymentID uint, paymentMethod string, paymentDate *time.Time, notes *string) (*models.Payment, error) {
	// Get existing payment
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, errors.DB(err, "failed to retrieve payment")
	}

	if payment == nil {
		return nil, errors.NotFound("payment", nil)
	}

	// Validate current status - can only confirm PENDING payments
	if payment.Status != models.PaymentStatusPending {
		return nil, errors.Validation(
			"Cannot confirm payment with status "+string(payment.Status)+", only PENDING payments can be confirmed",
			"status",
			string(payment.Status),
		)
	}

	// Validate payment method is not empty
	if paymentMethod == "" {
		return nil, errors.Validation(
			"Payment method is required",
			"payment_method",
			"empty",
		)
	}

	// Update payment status, date and payment method
	payment.Status = models.PaymentStatusPaid

	// Use provided date or current time
	if paymentDate != nil {
		payment.PaymentDate = paymentDate
	} else {
		now := time.Now()
		payment.PaymentDate = &now
	}

	payment.PaymentMethod = paymentMethod

	// Update notes if provided
	if notes != nil {
		payment.Notes = *notes
	}

	// Save to database
	err = s.paymentRepo.Update(ctx, payment)
	if err != nil {
		return nil, errors.DB(err, "failed to confirm payment")
	}

	// Crear entrada automática en cash_flows
	if err := s.createCashFlowForPayment(ctx, payment); err != nil {
		// Log error but don't fail the confirmation
		log.Printf("Warning: Failed to create cash flow entry for payment %d: %v", payment.ID, err)
	}

	return payment, nil
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

// GenerateAnnualFee crea una cuota anual para un año específico
func (s *paymentService) GenerateAnnualFee(ctx context.Context, year int, baseAmount float64) error {
	// Validar datos de entrada
	if baseAmount <= 0 {
		return errors.Validation("El monto base debe ser positivo", "baseAmount", "debe ser positivo")
	}

	// Verificar si ya existe una cuota para el año
	existingFee, err := s.membershipFeeRepo.FindByYear(ctx, year)
	if err != nil {
		return errors.DB(err, "error verificando cuota existente")
	}

	if existingFee != nil {
		return errors.New(errors.ErrDuplicateEntry, "ya existe una cuota para este año")
	}

	// Crear cuota anual
	fee := models.NewAnnualFee(year, baseAmount)

	return s.membershipFeeRepo.Create(ctx, fee)
}

// Deprecated: GenerateMonthlyFees - mantener por compatibilidad.
// Las cuotas ahora son anuales. Use GenerateAnnualFee en su lugar.
func (s *paymentService) GenerateMonthlyFees(ctx context.Context, year, _ int, baseAmount float64) error {
	// Simplemente delegar a GenerateAnnualFee ignorando el mes
	return s.GenerateAnnualFee(ctx, year, baseAmount)
}

func (s *paymentService) GetMembershipFee(ctx context.Context, year, _ int) (*models.MembershipFee, error) {
	// Las cuotas ahora son anuales, ignorar el mes
	fee, err := s.membershipFeeRepo.FindByYear(ctx, year)
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
		if p.PaymentDate != nil && (lastPaymentDate == nil || p.PaymentDate.After(*lastPaymentDate)) {
			lastPaymentDate = p.PaymentDate
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
		if p.PaymentDate != nil && (lastPaymentDate == nil || p.PaymentDate.After(*lastPaymentDate)) {
			lastPaymentDate = p.PaymentDate
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

	// Buscar todos los pagos pendientes (membership fees vencidas)
	now := time.Now()
	pendingStatus := models.PaymentStatusPending
	pendingPayments, err := s.paymentRepo.FindAll(ctx, &output.PaymentRepositoryFilters{
		Status: &pendingStatus,
	})
	if err != nil {
		return nil, errors.DB(err, "error buscando pagos pendientes")
	}

	// Agrupar por miembro y verificar si están en mora
	memberMap := make(map[uint]bool)
	for _, payment := range pendingPayments {
		// Solo procesar pagos de cuotas de membresía (no otros tipos de pagos)
		if payment.MembershipFeeID == nil {
			continue
		}

		// Verificar si la cuota está vencida
		if payment.MembershipFee != nil && !payment.MembershipFee.DueDate.Before(now) {
			continue
		}

		memberID := payment.MemberID
		if !memberMap[memberID] {
			statement, err := s.GetMemberStatement(ctx, memberID)
			if err != nil {
				// Loguear el error pero continuar con otros miembros
				continue
			}
			if statement.IsDefaulter {
				defaulters = append(defaulters, *statement)
				memberMap[memberID] = true
			}
		}
	}

	return defaulters, nil
}

func (s *paymentService) SendPaymentReminder(ctx context.Context, memberID uint) error {
	// Payment reminders are not implemented
	// If needed in the future, implement using EmailNotificationService (MailerSend)
	return errors.New(errors.ErrInvalidOperation, "payment reminders are not currently implemented")
}

func (s *paymentService) SendPaymentConfirmation(ctx context.Context, paymentID uint) error {
	// Payment confirmations are not implemented
	// If needed in the future, implement using EmailNotificationService (MailerSend)
	return errors.New(errors.ErrInvalidOperation, "payment confirmations are not currently implemented")
}

func (s *paymentService) SendDefaulterNotification(ctx context.Context, memberID uint, days int) error {
	// Defaulter notifications are not implemented
	// If needed in the future, implement using EmailNotificationService (MailerSend)
	return errors.New(errors.ErrInvalidOperation, "defaulter notifications are not currently implemented")
}

// ListPayments retrieves a paginated and filtered list of payments
func (s *paymentService) ListPayments(ctx context.Context, filters input.PaymentFilters) ([]*models.Payment, int, error) {
	// Validate pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}

	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // Maximum page size
	}

	// Set default ordering if not provided
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "payment_date DESC"
	}

	// Convert service filters to repository filters
	repoFilters := &output.PaymentRepositoryFilters{
		Status:        filters.Status,
		PaymentMethod: filters.PaymentMethod,
		StartDate:     filters.StartDate,
		EndDate:       filters.EndDate,
		MinAmount:     filters.MinAmount,
		MaxAmount:     filters.MaxAmount,
		MemberID:      filters.MemberID,
		Offset:        (page - 1) * pageSize,
		Limit:         pageSize,
		OrderBy:       orderBy,
	}

	// Get payments from repository
	payments, err := s.paymentRepo.FindAll(ctx, repoFilters)
	if err != nil {
		return nil, 0, errors.DB(err, "failed to list payments")
	}

	// Get total count for pagination
	total, err := s.paymentRepo.CountAll(ctx, repoFilters)
	if err != nil {
		return nil, 0, errors.DB(err, "failed to count payments")
	}

	// Convert to pointers
	result := make([]*models.Payment, len(payments))
	for i := range payments {
		result[i] = &payments[i]
	}

	return result, int(total), nil
}

// createCashFlowForPayment crea automáticamente una entrada en cash_flows cuando se confirma un pago
func (s *paymentService) createCashFlowForPayment(ctx context.Context, payment *models.Payment) error {
	// Verificar que el pago tiene una fecha de pago
	if payment.PaymentDate == nil {
		return errors.New(errors.ErrInvalidOperation, "payment date is required for cash flow creation")
	}

	// Verificar si ya existe un cash_flow para este pago (idempotencia)
	exists, err := s.cashFlowRepo.ExistsByPaymentID(ctx, payment.ID)
	if err != nil {
		return errors.DB(err, "failed to check existing cash flow")
	}

	if exists {
		log.Printf("Info: Cash flow already exists for payment %d, skipping creation", payment.ID)
		return nil
	}

	// Crear el cash flow usando el helper del modelo
	cashFlow, err := models.NewFromPayment(payment)
	if err != nil {
		return errors.Wrap(err, errors.ErrInvalidOperation, "failed to create cash flow from payment")
	}

	// Guardar en la base de datos
	if err := s.cashFlowRepo.Create(ctx, cashFlow); err != nil {
		return errors.DB(err, "failed to save cash flow entry")
	}

	log.Printf("Info: Cash flow entry created successfully for payment %d", payment.ID)
	return nil
}
