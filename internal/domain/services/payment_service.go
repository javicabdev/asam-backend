package services

import (
	"context"
	"fmt"
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
// Uses a database transaction to ensure Payment and CashFlow are created/updated atomically
func (s *paymentService) ConfirmPayment(ctx context.Context, paymentID uint, paymentMethod string, paymentDate *time.Time, notes *string, amount *float64) (*models.Payment, error) {
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

	// Update amount if provided and different from current amount
	if amount != nil && *amount != payment.Amount {
		// Validate amount is positive
		if *amount <= 0 {
			return nil, errors.Validation(
				"Amount must be greater than zero",
				"amount",
				fmt.Sprintf("%f", *amount),
			)
		}
		payment.Amount = *amount
	}

	// Update payment status, date and payment method
	payment.Status = models.PaymentStatusPaid

	// Use provided date or current time
	// If date is provided but has time 00:00:00, use current time instead
	if paymentDate != nil {
		// Check if the time component is midnight (00:00:00)
		// This likely means the frontend sent only a date without time
		if paymentDate.Hour() == 0 && paymentDate.Minute() == 0 && paymentDate.Second() == 0 {
			// Use the date from paymentDate but with the current time
			now := time.Now()
			combinedDate := time.Date(
				paymentDate.Year(), paymentDate.Month(), paymentDate.Day(),
				now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
				paymentDate.Location(),
			)
			payment.PaymentDate = &combinedDate
		} else {
			// Use the provided date and time as-is
			payment.PaymentDate = paymentDate
		}
	} else {
		now := time.Now()
		payment.PaymentDate = &now
	}

	payment.PaymentMethod = paymentMethod

	// Update notes if provided
	if notes != nil {
		payment.Notes = *notes
	}

	// Use transactional method to ensure Payment and CashFlow are updated/created atomically
	// This guarantees data consistency - both succeed or both fail (rollback)
	err = s.paymentRepo.ConfirmPaymentWithTransaction(ctx, payment)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "failed to confirm payment with transaction")
	}

	log.Printf("Payment %d confirmed successfully with synchronized cashflow", payment.ID)
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

// UpdatePaymentAndSyncCashFlow actualiza un payment y sincroniza su cashflow asociado en una transacción
func (s *paymentService) UpdatePaymentAndSyncCashFlow(ctx context.Context, payment *models.Payment) error {
	// Validar el pago
	if err := s.validatePayment(ctx, payment); err != nil {
		return err
	}

	// Delegar al repositorio que maneja la transacción
	return s.paymentRepo.UpdatePaymentAndSyncCashFlow(ctx, payment)
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

// GenerateAnnualFees genera cuotas anuales para todos los socios activos
func (s *paymentService) GenerateAnnualFees(ctx context.Context, req *input.GenerateAnnualFeesRequest) (*input.GenerateAnnualFeesResponse, error) {
	// Validar request
	if err := s.validateGenerateAnnualFeesRequest(req); err != nil {
		return nil, err
	}

	// 1. Crear o actualizar MembershipFee
	fee, err := s.ensureMembershipFee(ctx, req)
	if err != nil {
		return nil, err
	}

	// 2. Obtener todos los socios activos
	members, err := s.memberRepo.GetAllActive(ctx)
	if err != nil {
		return nil, errors.DB(err, "error obteniendo socios activos")
	}

	// 3. Generar pagos para cada socio
	response := &input.GenerateAnnualFeesResponse{
		Year:            req.Year,
		MembershipFeeID: fee.ID,
		TotalMembers:    len(members),
		Details:         make([]input.PaymentGenerationDetail, 0, len(members)),
	}

	for _, member := range members {
		detail := s.generatePaymentForMember(ctx, member, fee)
		response.Details = append(response.Details, detail)

		if detail.WasCreated {
			response.PaymentsGenerated++
			response.TotalExpectedAmount += detail.Amount
		} else if detail.Error == "" {
			response.PaymentsExisting++
		}
	}

	return response, nil
}

// validateGenerateAnnualFeesRequest valida el request de generación
func (s *paymentService) validateGenerateAnnualFeesRequest(req *input.GenerateAnnualFeesRequest) error {
	currentYear := time.Now().Year()

	if req.Year > currentYear {
		return errors.Validation(
			"No se pueden generar cuotas para años futuros",
			"year",
			"debe ser menor o igual al año actual",
		)
	}

	if req.Year < 2000 {
		return errors.Validation(
			"Año inválido",
			"year",
			"debe ser mayor o igual a 2000",
		)
	}

	if req.BaseFeeAmount <= 0 {
		return errors.Validation(
			"El monto base debe ser positivo",
			"baseFeeAmount",
			"debe ser mayor a 0",
		)
	}

	if req.FamilyFeeExtra < 0 {
		return errors.Validation(
			"El extra familiar debe ser no negativo",
			"familyFeeExtra",
			"debe ser mayor o igual a 0",
		)
	}

	return nil
}

// ensureMembershipFee crea o actualiza la cuota de membresía para el año
func (s *paymentService) ensureMembershipFee(ctx context.Context, req *input.GenerateAnnualFeesRequest) (*models.MembershipFee, error) {
	// Buscar cuota existente
	existingFee, err := s.membershipFeeRepo.FindByYear(ctx, req.Year)
	if err != nil {
		return nil, errors.DB(err, "error verificando cuota existente")
	}

	// Si ya existe, actualizarla
	if existingFee != nil {
		existingFee.BaseFeeAmount = req.BaseFeeAmount
		existingFee.FamilyFeeExtra = req.FamilyFeeExtra
		if err := s.membershipFeeRepo.Update(ctx, existingFee); err != nil {
			return nil, errors.DB(err, "error actualizando cuota existente")
		}
		return existingFee, nil
	}

	// Si no existe, crearla
	fee := models.NewAnnualFee(req.Year, req.BaseFeeAmount)
	fee.FamilyFeeExtra = req.FamilyFeeExtra

	if err := s.membershipFeeRepo.Create(ctx, fee); err != nil {
		return nil, errors.DB(err, "error creando cuota anual")
	}

	return fee, nil
}

// generatePaymentForMember genera un pago pendiente para un socio específico
func (s *paymentService) generatePaymentForMember(ctx context.Context, member *models.Member, fee *models.MembershipFee) input.PaymentGenerationDetail {
	detail := input.PaymentGenerationDetail{
		MemberID:     member.ID,
		MemberNumber: member.MembershipNumber,
		MemberName:   member.NombreCompleto(),
	}

	// Calcular monto según tipo de membresía
	isFamily := member.IsFamiliar()
	detail.Amount = fee.Calculate(isFamily)

	// Verificar si el año de la cuota es anterior al año de alta del socio
	memberRegistrationYear := member.RegistrationDate.Year()
	if fee.Year < memberRegistrationYear {
		detail.WasCreated = false
		detail.Error = ""
		return detail
	}

	// Verificar si ya existe un pago para este socio y año
	// Buscamos pagos del año completo
	yearStart := time.Date(fee.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(fee.Year, 12, 31, 23, 59, 59, 0, time.UTC)

	existingPayments, err := s.paymentRepo.FindByMember(ctx, member.ID, yearStart, yearEnd)
	if err != nil {
		detail.Error = "error verificando pagos existentes: " + err.Error()
		return detail
	}

	// Verificar si ya existe un pago asociado a esta cuota
	for _, payment := range existingPayments {
		if payment.MembershipFeeID != nil && *payment.MembershipFeeID == fee.ID {
			detail.WasCreated = false
			return detail
		}
	}

	// Crear pago pendiente
	feeID := fee.ID
	payment := &models.Payment{
		MemberID:        member.ID,
		Amount:          detail.Amount,
		Status:          models.PaymentStatusPending,
		MembershipFeeID: &feeID,
		PaymentDate:     nil,
		PaymentMethod:   "",
		Notes:           "Cuota anual generada automáticamente",
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		detail.Error = "error creando pago: " + err.Error()
		return detail
	}

	detail.WasCreated = true
	return detail
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

func (s *paymentService) ListMembershipFees(ctx context.Context, page, pageSize int) ([]*models.MembershipFee, int, error) {
	// Validar paginación
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Calcular offset
	offset := (page - 1) * pageSize

	// Obtener cuotas del repositorio
	fees, total, err := s.membershipFeeRepo.FindAll(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, errors.DB(err, "error listando cuotas de membresía")
	}

	// Convertir a slice de punteros
	result := make([]*models.MembershipFee, len(fees))
	for i := range fees {
		result[i] = &fees[i]
	}

	return result, int(total), nil
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
	// Usar el método optimizado del repositorio que obtiene todos los datos en 3 queries
	// en lugar de N+1 queries (una por cada miembro moroso)
	defaultersData, err := s.paymentRepo.GetDefaultersData(ctx)
	if err != nil {
		return nil, errors.DB(err, "error obteniendo datos de morosos")
	}

	// Convertir DefaulterData a AccountStatement
	defaulters := make([]input.AccountStatement, len(defaultersData))
	now := time.Now()

	for i, data := range defaultersData {
		// Calcular días de mora basado en la cuota más antigua vencida
		defaultDays := 0
		if !data.OldestPendingDue.IsZero() && now.After(data.OldestPendingDue) {
			defaultDays = int(now.Sub(data.OldestPendingDue).Hours() / 24)
		}

		defaulters[i] = input.AccountStatement{
			TotalPaid:       data.TotalPaid,
			PendingPayments: data.PendingPayments,
			PaymentHistory:  data.PaymentHistory,
			LastPaymentDate: data.LastPaymentDate,
			IsDefaulter:     data.OverdueCount > 0,
			DefaultDays:     defaultDays,
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
