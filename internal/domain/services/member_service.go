package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"github.com/javicabdev/asam-backend/pkg/logger/audit"
	"github.com/javicabdev/asam-backend/pkg/metrics"
)

type memberService struct {
	repository        output.MemberRepository
	paymentRepository output.PaymentRepository
	membershipFeeRepo output.MembershipFeeRepository
	appLogger         logger.Logger
	auditLogger       audit.Logger
}

// NewMemberService crea una nueva instancia del servicio de miembros
func NewMemberService(
	repository output.MemberRepository,
	paymentRepository output.PaymentRepository,
	membershipFeeRepo output.MembershipFeeRepository,
	appLogger logger.Logger,
	auditLogger audit.Logger,
) input.MemberService {
	return &memberService{
		repository:        repository,
		paymentRepository: paymentRepository,
		membershipFeeRepo: membershipFeeRepo,
		appLogger:         appLogger,
		auditLogger:       auditLogger,
	}
}

// CreateMember implementa la lógica de creación de un nuevo miembro
func (s *memberService) CreateMember(ctx context.Context, member *models.Member) error {
	// Logging al inicio de la operación
	s.appLogger.Info("Creating new member",
		zap.String("numero_socio", member.MembershipNumber),
		zap.String("tipo_membresia", member.MembershipType))

	// ========================================
	// PHASE 1: Pre-transaction validations
	// ========================================

	// Validar duplicados
	if err := s.validateDuplicateMember(ctx, member); err != nil {
		return err
	}

	// Establecer valores por defecto
	s.setDefaultValues(member)

	// Validar el miembro antes de crear
	if err := s.validateMember(ctx, member); err != nil {
		return err
	}

	// ========================================
	// PHASE 2: Atomic transaction
	// ========================================

	if err := s.createMemberWithPayment(ctx, member); err != nil {
		return err
	}

	// ========================================
	// PHASE 3: Post-transaction operations
	// ========================================

	s.updateMetrics(member)
	s.logSuccess(ctx, member)

	return nil
}

// GetMemberByID obtiene un miembro por su ID
func (s *memberService) GetMemberByID(ctx context.Context, id uint) (*models.Member, error) {
	member, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.appLogger.Error("Error getting member by ID",
			zap.Uint("id", id),
			zap.Error(err))
		return nil, errors.DB(err, "error obteniendo miembro por ID")
	}

	if member == nil {
		return nil, errors.NotFound("member", nil)
	}

	return member, nil
}

// GetMemberByNumeroSocio obtiene un miembro por su número de socio
func (s *memberService) GetMemberByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	member, err := s.repository.GetByNumeroSocio(ctx, numeroSocio)
	if err != nil {
		s.appLogger.Error("Error getting member by numero socio",
			zap.String("numero_socio", numeroSocio),
			zap.Error(err))
		return nil, errors.DB(err, "error obteniendo miembro por numero socio")
	}

	if member == nil {
		return nil, errors.NotFound("member", nil)
	}

	return member, nil
}

// UpdateMember actualiza los datos de un miembro existente
func (s *memberService) UpdateMember(ctx context.Context, member *models.Member) error {
	// Verificar que el miembro existe
	existing, err := s.repository.GetByID(ctx, member.ID)
	if err != nil {
		s.appLogger.Error("Error checking existing member",
			zap.Uint("id", member.ID),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(member.ID),
			"Error al verificar existencia del miembro", err)
		return errors.DB(err, "error verificando miembro existente")
	}

	if existing == nil {
		s.appLogger.Error("Member not found", zap.Uint("id", member.ID))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(member.ID),
			"Miembro no encontrado", nil)
		return errors.NotFound("member", nil)
	}

	// No permitir cambios en campos inmutables
	member.MembershipNumber = existing.MembershipNumber
	member.RegistrationDate = existing.RegistrationDate

	// Validar el miembro antes de actualizar
	if validateErr := member.Validate(); validateErr != nil {
		s.appLogger.Error("Member validation failed",
			zap.Uint("id", member.ID),
			zap.Error(validateErr))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(member.ID),
			"Error en la validación del miembro", validateErr)

		// Conservar el error de validación si ya es un AppError, sino convertirlo
		appErr, ok := errors.AsAppError(validateErr)
		if ok {
			return appErr
		}
		return errors.Validation("Error validando miembro", "", validateErr.Error())
	}

	// Actualizar el miembro
	if err = s.repository.Update(ctx, member); err != nil {
		s.appLogger.Error("Failed to update member",
			zap.Uint("id", member.ID),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(member.ID),
			"Error al actualizar miembro en base de datos", err)
		return errors.DB(err, "error actualizando miembro")
	}

	// Log de auditoría con los cambios
	s.auditLogger.LogChange(ctx, audit.ActionUpdate, audit.EntityMember,
		numToStr(member.ID),
		existing, // datos anteriores
		member,   // datos nuevos
		"Updated member with numero_socio "+member.MembershipNumber)

	s.appLogger.Info("Member updated successfully",
		zap.String("numero_socio", member.MembershipNumber),
		zap.Uint("member_id", member.ID))

	return nil
}

// DeactivateMember implementa la lógica de baja de un miembro
func (s *memberService) DeactivateMember(ctx context.Context, id uint, fechaBaja *time.Time) error {
	// Obtener el miembro
	member, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.appLogger.Error("Error getting member",
			zap.Uint("id", id),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Error al obtener miembro para desactivación", err)
		return errors.DB(err, "error obteniendo miembro")
	}

	if member == nil {
		s.appLogger.Error("Member not found", zap.Uint("id", id))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Miembro no encontrado", nil)
		return errors.NotFound("member", nil)
	}

	// Guardar estado anterior para métricas
	previousStatus := member.State
	previousType := member.MembershipType

	// Verificar que no esté ya inactivo
	if member.State == models.EstadoInactivo {
		s.appLogger.Warn("Member already inactive", zap.Uint("id", id))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Intento de desactivar miembro ya inactivo", nil)
		return errors.New(errors.ErrInvalidOperation, "el miembro ya está dado de baja")
	}

	// Verificar si el miembro tiene pagos pendientes
	hasPendingPayments, err := s.paymentRepository.HasPendingPayments(ctx, id)
	if err != nil {
		s.appLogger.Error("Error checking pending payments",
			zap.Uint("id", id),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Error al verificar pagos pendientes", err)
		return errors.DB(err, "error verificando pagos pendientes")
	}

	if hasPendingPayments {
		s.appLogger.Warn("Cannot deactivate member with pending payments",
			zap.Uint("id", id),
			zap.String("numero_socio", member.MembershipNumber))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Intento de desactivar miembro con pagos pendientes", nil)
		return errors.New(errors.ErrInvalidOperation, "no se puede dar de baja un socio con pagos pendientes")
	}

	// Guardar estado anterior para el log de auditoría
	previousState := *member

	// Establecer fecha de baja
	if fechaBaja == nil {
		now := time.Now()
		fechaBaja = &now
	}
	member.LeavingDate = fechaBaja
	member.State = models.EstadoInactivo

	// Validar y guardar cambios
	if err := member.Validate(); err != nil {
		s.appLogger.Error("Member validation failed",
			zap.Uint("id", id),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Error en la validación del miembro", err)

		appErr, ok := errors.AsAppError(err)
		if ok {
			return appErr
		}
		return errors.Validation("Error validando miembro", "", err.Error())
	}

	if err := s.repository.Update(ctx, member); err != nil {
		s.appLogger.Error("Failed to deactivate member",
			zap.Uint("id", id),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Error al desactivar miembro en base de datos", err)
		return errors.DB(err, "error desactivando miembro")
	}

	// Log de auditoría con los cambios
	s.auditLogger.LogChange(ctx, audit.ActionUpdate, audit.EntityMember,
		numToStr(id),
		&previousState,
		member,
		"Deactivated member with numero_socio "+member.MembershipNumber)

	s.appLogger.Info("Member deactivated successfully",
		zap.String("numero_socio", member.MembershipNumber),
		zap.Uint("member_id", member.ID))

	// Actualizar métricas
	metrics.MembersByStatus.WithLabelValues(
		previousStatus,
		previousType,
	).Dec()

	metrics.MembersByStatus.WithLabelValues(
		member.State,
		member.MembershipType,
	).Inc()

	return nil
}

// ListMembers obtiene una lista de miembros según los criterios especificados
func (s *memberService) ListMembers(ctx context.Context, filters input.MemberFilters) ([]*models.Member, error) {
	// Convertir filtros de input a output
	repoFilters := output.MemberFilters{
		Estado:        filters.State,
		TipoMembresia: filters.MembershipType,
		SearchTerm:    filters.SearchTerm,
		Page:          filters.Page,
		PageSize:      filters.PageSize,
		OrderBy:       filters.OrderBy,
	}

	members, err := s.repository.List(ctx, repoFilters)
	if err != nil {
		s.appLogger.Error("Error listing members", zap.Error(err))
		return nil, errors.DB(err, "error al listar miembros")
	}

	// Convertir []models.Member a []*models.Member
	result := make([]*models.Member, len(members))
	for i := range members {
		result[i] = &members[i]
	}

	return result, nil
}

// validateDuplicateMember verifica si ya existe un miembro con el mismo número o documento
func (s *memberService) validateDuplicateMember(ctx context.Context, member *models.Member) error {
	// Verificar si ya existe un miembro con el mismo número de socio
	existing, err := s.repository.GetByNumeroSocio(ctx, member.MembershipNumber)
	if err != nil {
		s.appLogger.Error("Error checking existing member",
			zap.String("numero_socio", member.MembershipNumber),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Error al verificar miembro existente", err)
		return errors.DB(err, "error verificando miembro existente")
	}

	if existing != nil {
		validationErr := errors.NewValidationError(
			"El número de socio ya está registrado",
			map[string]string{
				"numero_socio": fmt.Sprintf("Ya existe un miembro con el número de socio %s", member.MembershipNumber),
			},
		)
		s.appLogger.Warn("Attempted to create duplicate member",
			zap.String("numero_socio", member.MembershipNumber))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Intento de crear miembro duplicado", validationErr)
		return validationErr
	}

	// Verificar si ya existe un miembro con el mismo documento de identidad
	if member.IdentityCard != nil && *member.IdentityCard != "" {
		return s.validateDuplicateIdentityCard(ctx, member)
	}

	return nil
}

// validateDuplicateIdentityCard verifica si ya existe un miembro con el mismo documento
func (s *memberService) validateDuplicateIdentityCard(ctx context.Context, member *models.Member) error {
	existingByDNI, err := s.repository.GetByIdentityCard(ctx, *member.IdentityCard)
	if err != nil {
		s.appLogger.Error("Error checking existing member by identity card",
			zap.String("identity_card", *member.IdentityCard),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Error al verificar documento de identidad", err)
		return errors.DB(err, "error verificando documento de identidad")
	}

	if existingByDNI != nil {
		validationErr := errors.NewValidationError(
			"El documento de identidad ya está registrado",
			map[string]string{
				"documento_identidad": fmt.Sprintf("Ya existe un miembro (%s) con este documento de identidad", existingByDNI.MembershipNumber),
			},
		)
		s.appLogger.Warn("Attempted to create member with duplicate identity card",
			zap.String("identity_card", *member.IdentityCard),
			zap.String("existing_member", existingByDNI.MembershipNumber))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Intento de crear miembro con documento de identidad duplicado", validationErr)
		return validationErr
	}

	return nil
}

// setDefaultValues establece los valores por defecto para un nuevo miembro
func (s *memberService) setDefaultValues(member *models.Member) {
	if member.State == "" {
		member.State = models.EstadoActivo
	}
	if member.Province == "" {
		member.Province = "Barcelona"
	}
	if member.Country == "" {
		member.Country = "España"
	}
	if member.Nationality == "" {
		member.Nationality = "Senegal"
	}
}

// validateMember valida el miembro y convierte errores a AppError
func (s *memberService) validateMember(ctx context.Context, member *models.Member) error {
	if err := member.Validate(); err != nil {
		s.appLogger.Error("Member validation failed",
			zap.String("numero_socio", member.MembershipNumber),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Error en la validación del miembro", err)

		// Conservar el error de validación si ya es un AppError, sino convertirlo
		appErr, ok := errors.AsAppError(err)
		if ok {
			return appErr
		}
		return errors.Validation("Error validando miembro", "", err.Error())
	}
	return nil
}

// createMemberWithPayment crea el miembro y el pago pendiente en una transacción
func (s *memberService) createMemberWithPayment(ctx context.Context, member *models.Member) error {
	// Begin transaction
	tx, err := s.repository.BeginTransaction(ctx)
	if err != nil {
		s.appLogger.Error("Failed to begin transaction",
			zap.String("numero_socio", member.MembershipNumber),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Error al iniciar transacción", err)
		return errors.DB(err, "error iniciando transacción")
	}

	// Ensure rollback on error
	defer func() {
		if err != nil {
			s.rollbackTransaction(tx, member.MembershipNumber)
		}
	}()

	// Create the member in the database within transaction
	if err = s.repository.CreateWithTx(ctx, tx, member); err != nil {
		s.appLogger.Error("Failed to create member",
			zap.String("numero_socio", member.MembershipNumber),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Error al crear miembro en base de datos", err)
		return errors.DB(err, "error creando miembro")
	}

	// Create the pending payment for the current year
	if err = s.createPendingPayment(ctx, tx, member); err != nil {
		s.appLogger.Error("Failed to create pending payment",
			zap.String("numero_socio", member.MembershipNumber),
			zap.Uint("member_id", member.ID),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Error al crear pago pendiente", err)
		return err // Already wrapped in createPendingPayment
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		s.appLogger.Error("Failed to commit transaction",
			zap.String("numero_socio", member.MembershipNumber),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Error al confirmar transacción", err)
		return errors.DB(err, "error confirmando transacción")
	}

	return nil
}

// rollbackTransaction ejecuta rollback de una transacción y loguea errores
func (s *memberService) rollbackTransaction(tx output.Transaction, membershipNumber string) {
	if rbErr := tx.Rollback(); rbErr != nil {
		s.appLogger.Error("Failed to rollback transaction",
			zap.String("numero_socio", membershipNumber),
			zap.Error(rbErr))
	}
}

// updateMetrics actualiza las métricas de miembros
func (s *memberService) updateMetrics(member *models.Member) {
	metrics.MembersByStatus.WithLabelValues(
		member.State,
		member.MembershipType,
	).Inc()
}

// logSuccess registra el éxito de la creación del miembro
func (s *memberService) logSuccess(ctx context.Context, member *models.Member) {
	// Registrar la acción en el log de auditoría
	s.auditLogger.LogAction(ctx,
		audit.ActionCreate,
		audit.EntityMember,
		member.MembershipNumber,
		"Created new member with pending payment")

	s.appLogger.Info("Member created successfully with pending payment",
		zap.String("numero_socio", member.MembershipNumber),
		zap.Uint("member_id", member.ID))
}

// createPendingPayment creates a pending payment for a newly created member
func (s *memberService) createPendingPayment(ctx context.Context, tx output.Transaction, member *models.Member) error {
	currentYear := time.Now().Year()

	// Get the annual membership fee for the current year
	fee, err := s.membershipFeeRepo.FindByYearWithTx(ctx, tx, currentYear)
	if err != nil {
		s.appLogger.Error("Error fetching membership fee",
			zap.Int("year", currentYear),
			zap.Error(err))
		return errors.DB(err, "error obteniendo cuota de socio")
	}

	if fee == nil {
		s.appLogger.Warn("No membership fee found for current year",
			zap.Int("year", currentYear))
		return errors.New(errors.ErrNotFound, fmt.Sprintf("no existe cuota para el año %d", currentYear))
	}

	// Determine payment amount based on membership type
	isFamily := member.MembershipType == models.TipoMembresiaPFamiliar
	amount := fee.Calculate(isFamily)

	// Create the pending payment
	payment := &models.Payment{
		MemberID:        member.ID,
		Amount:          amount,
		PaymentDate:     nil, // Nil indicates pending payment
		Status:          models.PaymentStatusPending,
		PaymentMethod:   "",
		Notes:           "",
		MembershipFeeID: &fee.ID,
	}

	if err := payment.Validate(); err != nil {
		s.appLogger.Error("Payment validation failed",
			zap.Uint("member_id", member.ID),
			zap.Error(err))
		return errors.Validation("Error validando pago", "", err.Error())
	}

	if err := s.paymentRepository.CreateWithTx(ctx, tx, payment); err != nil {
		s.appLogger.Error("Failed to create pending payment",
			zap.Uint("member_id", member.ID),
			zap.Error(err))
		return errors.DB(err, "error creando pago pendiente")
	}

	s.appLogger.Info("Pending payment created successfully",
		zap.Uint("member_id", member.ID),
		zap.Uint("payment_id", payment.ID),
		zap.Float64("amount", payment.Amount))

	return nil
}

// numToStr es una función auxiliar para convertir un número a string para los logs
func numToStr(num uint) string {
	return strconv.FormatUint(uint64(num), 10)
}

// GetNextMemberNumber obtiene el siguiente número de socio disponible según el tipo
func (s *memberService) GetNextMemberNumber(ctx context.Context, isFamily bool) (string, error) {
	// Determinar el prefijo según el tipo
	prefix := "B" // Individual por defecto
	if isFamily {
		prefix = "A" // Familiar
	}

	// Obtener el último número con ese prefijo
	lastNumber, err := s.repository.GetLastMemberNumberByPrefix(ctx, prefix)
	if err != nil {
		s.appLogger.Error("Error getting last member number",
			zap.String("prefix", prefix),
			zap.Error(err))
		return "", errors.DB(err, "error al obtener el último número de socio")
	}

	var nextNum int
	if lastNumber == "" {
		// No hay ningún socio con ese prefijo, empezar desde 1
		nextNum = 1
		s.appLogger.Info("No members found with prefix, starting from 1",
			zap.String("prefix", prefix))
	} else {
		// Extraer la parte numérica y incrementar
		numPart := lastNumber[1:] // Quitar el prefijo
		currentNum, parseErr := strconv.Atoi(numPart)
		if parseErr != nil {
			s.appLogger.Error("Error parsing member number",
				zap.String("lastNumber", lastNumber),
				zap.Error(parseErr))
			return "", errors.NewInternalError("error al procesar el número de socio")
		}
		nextNum = currentNum + 1
	}

	// Formatear con padding de al menos 5 dígitos, pero soportar más si es necesario
	var formattedNumber string
	if nextNum < 100000 {
		// Usar padding de 5 dígitos para números menores a 100000
		formattedNumber = fmt.Sprintf("%s%05d", prefix, nextNum)
	} else {
		// Para números mayores, no aplicar padding adicional
		formattedNumber = fmt.Sprintf("%s%d", prefix, nextNum)
	}

	s.appLogger.Info("Generated next member number",
		zap.String("prefix", prefix),
		zap.String("nextNumber", formattedNumber))

	return formattedNumber, nil
}

// CheckMemberNumberExists verifica si un número de socio ya existe
func (s *memberService) CheckMemberNumberExists(ctx context.Context, memberNumber string) (bool, error) {
	// Validar el formato del número de socio
	if !isValidMemberNumber(memberNumber) {
		return false, errors.NewValidationError(
			"Formato de número de socio inválido",
			map[string]string{"memberNumber": "Debe seguir el formato [A|B] seguido de al menos 5 dígitos"},
		)
	}

	member, err := s.repository.GetByNumeroSocio(ctx, memberNumber)
	if err != nil {
		s.appLogger.Error("Error checking member number existence",
			zap.String("memberNumber", memberNumber),
			zap.Error(err))
		return false, errors.DB(err, "error al verificar el número de socio")
	}

	// Si member es nil, el número no existe
	exists := member != nil

	s.appLogger.Debug("Checked member number existence",
		zap.String("memberNumber", memberNumber),
		zap.Bool("exists", exists))

	return exists, nil
}

// isValidMemberNumber valida el formato del número de socio
func isValidMemberNumber(memberNumber string) bool {
	// El número debe empezar con A o B seguido de al menos 5 dígitos
	if len(memberNumber) < 6 {
		return false
	}

	prefix := memberNumber[0]
	if prefix != 'A' && prefix != 'B' {
		return false
	}

	// Verificar que el resto sean dígitos
	for i := 1; i < len(memberNumber); i++ {
		if memberNumber[i] < '0' || memberNumber[i] > '9' {
			return false
		}
	}

	return true
}

// SearchMembersWithoutUser busca miembros que no tienen usuario asociado
func (s *memberService) SearchMembersWithoutUser(ctx context.Context, criteria string) ([]*models.Member, error) {
	// Validar el criterio de búsqueda
	if strings.TrimSpace(criteria) == "" {
		return nil, errors.NewValidationError(
			"Criterio de búsqueda vacío",
			map[string]string{"criteria": "El criterio de búsqueda no puede estar vacío"},
		)
	}

	// Si el criterio es muy corto, podría generar demasiados resultados
	if len(strings.TrimSpace(criteria)) < 2 {
		return nil, errors.NewValidationError(
			"Criterio de búsqueda muy corto",
			map[string]string{"criteria": "El criterio debe tener al menos 2 caracteres"},
		)
	}

	s.appLogger.Debug("Searching members without user",
		zap.String("criteria", criteria))

	// Llamar al repositorio para buscar miembros sin usuario
	members, err := s.repository.SearchWithoutUser(ctx, criteria)
	if err != nil {
		s.appLogger.Error("Error searching members without user",
			zap.String("criteria", criteria),
			zap.Error(err))
		return nil, errors.DB(err, "error al buscar miembros sin usuario")
	}

	// Convertir []models.Member a []*models.Member
	result := make([]*models.Member, len(members))
	for i := range members {
		result[i] = &members[i]
	}

	s.appLogger.Info("Members without user found",
		zap.String("criteria", criteria),
		zap.Int("count", len(result)))

	return result, nil
}
