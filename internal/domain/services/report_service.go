package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

const (
	debtorTypeIndividual = "INDIVIDUAL"
	debtorTypeFamily     = "FAMILY"

	// Valores del campo MembershipType en la BD (minúsculas)
	membershipTypeIndividual = "individual"
	membershipTypeFamily     = "familiar"
)

type reportService struct {
	reportRepo output.ReportRepository
	memberRepo output.MemberRepository
	familyRepo output.FamilyRepository
}

// NewReportService crea una nueva instancia del servicio de reportes
func NewReportService(
	reportRepo output.ReportRepository,
	memberRepo output.MemberRepository,
	familyRepo output.FamilyRepository,
) input.ReportService {
	return &reportService{
		reportRepo: reportRepo,
		memberRepo: memberRepo,
		familyRepo: familyRepo,
	}
}

// GetDelinquentReport genera el reporte de morosos
func (s *reportService) GetDelinquentReport(ctx context.Context, inputParams input.DelinquentReportInput) (*input.DelinquentReportResponse, error) {
	// 1. Establecer fecha de corte
	cutoffDate := s.getCutoffDate(inputParams)

	// 2. Obtener todos los pagos PENDIENTES
	pendingPayments, err := s.reportRepo.GetPendingPayments(ctx)
	if err != nil {
		return nil, errors.DB(err, "error obteniendo pagos pendientes")
	}

	// 3. Agrupar pagos por deudor
	debtorMap := s.buildDebtorMap(ctx, pendingPayments, cutoffDate)

	// 4. Convertir map a slice y aplicar filtros
	debtors := s.processDebtors(debtorMap, inputParams)

	// 5. Calcular resumen estadístico
	summary := s.calculateSummary(debtors)
	totalCount := len(debtors)

	// 6. Aplicar paginación
	debtors = s.paginateDebtors(debtors, inputParams)

	// 7. Retornar respuesta
	return &input.DelinquentReportResponse{
		Debtors:     debtors,
		TotalCount:  totalCount,
		Summary:     summary,
		GeneratedAt: time.Now(),
	}, nil
}

// getCutoffDate obtiene la fecha de corte para el reporte
func (s *reportService) getCutoffDate(inputParams input.DelinquentReportInput) time.Time {
	if inputParams.CutoffDate != nil {
		return *inputParams.CutoffDate
	}
	return time.Now()
}

// buildDebtorMap agrupa pagos por deudor
func (s *reportService) buildDebtorMap(ctx context.Context, pendingPayments []models.Payment, cutoffDate time.Time) map[string]*input.Debtor {
	debtorMap := make(map[string]*input.Debtor)

	for i := range pendingPayments {
		payment := &pendingPayments[i]

		if payment.MemberID == 0 {
			continue // Skip pagos sin member_id
		}

		debtorKey := fmt.Sprintf("member_%d", payment.MemberID)

		// Crear deudor si no existe
		if _, exists := debtorMap[debtorKey]; !exists {
			debtor, err := s.createDebtor(ctx, payment.MemberID)
			if err != nil || debtor == nil {
				continue
			}
			debtorMap[debtorKey] = debtor
		}

		// Añadir pago al deudor
		s.addPaymentToDebtor(debtorMap[debtorKey], payment, cutoffDate)
	}

	return debtorMap
}

// addPaymentToDebtor añade un pago pendiente a un deudor
func (s *reportService) addPaymentToDebtor(debtor *input.Debtor, payment *models.Payment, cutoffDate time.Time) {
	daysOverdue := int(cutoffDate.Sub(payment.CreatedAt).Hours() / 24)

	pendingPayment := &input.PendingPayment{
		ID:          payment.ID,
		Amount:      payment.Amount,
		CreatedAt:   payment.CreatedAt,
		DaysOverdue: daysOverdue,
	}
	if payment.Notes != "" {
		pendingPayment.Notes = &payment.Notes
	}

	debtor.PendingPayments = append(debtor.PendingPayments, pendingPayment)
	debtor.TotalDebt += payment.Amount

	// Actualizar el pago más antiguo
	if debtor.OldestDebtDate.IsZero() || payment.CreatedAt.Before(debtor.OldestDebtDate) {
		debtor.OldestDebtDate = payment.CreatedAt
		debtor.OldestDebtDays = daysOverdue
	}
}

// processDebtors convierte el map a slice y aplica filtros y ordenamiento
func (s *reportService) processDebtors(debtorMap map[string]*input.Debtor, inputParams input.DelinquentReportInput) []*input.Debtor {
	// Convertir map a slice
	debtors := make([]*input.Debtor, 0, len(debtorMap))
	for _, debtor := range debtorMap {
		debtors = append(debtors, debtor)
	}

	// Aplicar filtros
	debtors = s.applyFilters(debtors, inputParams)

	// Ordenar
	debtors = s.sortDebtors(debtors, inputParams.SortBy)

	return debtors
}

// paginateDebtors aplica paginación a la lista de deudores
func (s *reportService) paginateDebtors(debtors []*input.Debtor, inputParams input.DelinquentReportInput) []*input.Debtor {
	page := inputParams.Page
	pageSize := inputParams.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx > len(debtors) {
		return []*input.Debtor{}
	}

	if endIdx > len(debtors) {
		endIdx = len(debtors)
	}

	return debtors[startIdx:endIdx]
}

// createDebtor crea un nuevo deudor con su información
func (s *reportService) createDebtor(ctx context.Context, memberID uint) (*input.Debtor, error) {
	// Cargar información del socio
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}

	// Determinar el tipo real basado en el campo MembershipType de la BD
	actualDebtorType := debtorTypeIndividual
	isFamily := member.MembershipType == membershipTypeFamily

	if isFamily {
		actualDebtorType = debtorTypeFamily
	}

	debtor := &input.Debtor{
		Type:            actualDebtorType,
		PendingPayments: []*input.PendingPayment{},
		TotalDebt:       0,
	}

	// Si es FAMILY, construir objeto family y dejar member en null
	if isFamily {
		// Obtener la familia del miembro origen
		family, err := s.familyRepo.GetByOriginMemberID(ctx, memberID)
		if err != nil {
			return nil, err
		}

		if family != nil {
			familyInfo := s.loadFamilyInfo(ctx, family, member)

			familyID := family.ID
			debtor.FamilyID = &familyID
			debtor.Family = familyInfo
			debtor.MemberID = nil
			debtor.Member = nil
		}
	} else {
		// Si es INDIVIDUAL, construir objeto member y dejar family en null
		memberInfo, err := s.loadMemberInfo(ctx, memberID)
		if err != nil {
			return nil, err
		}

		debtor.MemberID = &memberID
		debtor.Member = memberInfo
		debtor.FamilyID = nil
		debtor.Family = nil
	}

	// Obtener último pago PAID de este socio
	lastPaid, err := s.reportRepo.GetLastPaidPaymentForMember(ctx, memberID)
	if err == nil && lastPaid != nil {
		debtor.LastPaymentDate = &lastPaid.PaymentDate
		debtor.LastPaymentAmount = &lastPaid.Amount
	}

	return debtor, nil
}

// loadFamilyInfo carga la información de una familia
func (s *reportService) loadFamilyInfo(ctx context.Context, family *models.Family, primaryMember *models.Member) *input.DebtorFamilyInfo {
	// Obtener familiares para contar total de miembros
	familiares, err := s.familyRepo.GetFamiliares(ctx, family.ID)
	if err != nil {
		// Si hay error, asumir solo el miembro principal
		familiares = []*models.Familiar{}
	}

	// Total de miembros = miembro principal + familiares + cónyuges (si existen)
	totalMembers := 1 + len(familiares)
	if family.EsposoNombre != "" || family.EsposaNombre != "" {
		totalMembers++
	}

	// Construir nombre de familia
	familyName := fmt.Sprintf("Familia %s", primaryMember.Surnames)
	if family.EsposoApellidos != "" {
		familyName = fmt.Sprintf("Familia %s", family.EsposoApellidos)
	}

	// Construir información del miembro principal
	primaryMemberInfo := &input.DebtorMemberInfo{
		ID:           primaryMember.ID,
		MemberNumber: primaryMember.MembershipNumber,
		FirstName:    primaryMember.Name,
		LastName:     primaryMember.Surnames,
		Status:       primaryMember.State,
	}

	if primaryMember.Email != nil && *primaryMember.Email != "" {
		primaryMemberInfo.Email = primaryMember.Email
	}

	// Obtener primer teléfono si existe
	if len(primaryMember.Telefonos) > 0 {
		phone := primaryMember.Telefonos[0].NumeroTelefono
		primaryMemberInfo.Phone = &phone
	}

	familyInfo := &input.DebtorFamilyInfo{
		ID:            family.ID,
		FamilyName:    familyName,
		PrimaryMember: *primaryMemberInfo,
		TotalMembers:  totalMembers,
	}

	return familyInfo
}

// loadMemberInfo carga la información de un socio
func (s *reportService) loadMemberInfo(ctx context.Context, memberID uint) (*input.DebtorMemberInfo, error) {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.NotFound("member", nil)
	}

	memberInfo := &input.DebtorMemberInfo{
		ID:           member.ID,
		MemberNumber: member.MembershipNumber,
		FirstName:    member.Name,
		LastName:     member.Surnames,
		Status:       member.State,
	}

	if member.Email != nil && *member.Email != "" {
		memberInfo.Email = member.Email
	}

	// Obtener primer teléfono si existe
	if len(member.Telefonos) > 0 {
		phone := member.Telefonos[0].NumeroTelefono
		memberInfo.Phone = &phone
	}

	return memberInfo, nil
}

// applyFilters aplica los filtros según input
func (s *reportService) applyFilters(debtors []*input.Debtor, inputParams input.DelinquentReportInput) []*input.Debtor {
	filtered := make([]*input.Debtor, 0)

	for _, debtor := range debtors {
		// Filtro por importe mínimo
		if inputParams.MinAmount != nil && debtor.TotalDebt < *inputParams.MinAmount {
			continue
		}

		// Filtro por tipo de deudor
		if inputParams.DebtorType != nil && debtor.Type != *inputParams.DebtorType {
			continue
		}

		filtered = append(filtered, debtor)
	}

	return filtered
}

// sortDebtors ordena deudores según criterio
func (s *reportService) sortDebtors(debtors []*input.Debtor, sortBy *string) []*input.Debtor {
	sortCriteria := "DAYS_DESC" // default
	if sortBy != nil && *sortBy != "" {
		sortCriteria = *sortBy
	}

	switch sortCriteria {
	case "AMOUNT_DESC":
		sort.Slice(debtors, func(i, j int) bool {
			return debtors[i].TotalDebt > debtors[j].TotalDebt
		})
	case "AMOUNT_ASC":
		sort.Slice(debtors, func(i, j int) bool {
			return debtors[i].TotalDebt < debtors[j].TotalDebt
		})
	case "DAYS_ASC":
		sort.Slice(debtors, func(i, j int) bool {
			return debtors[i].OldestDebtDays < debtors[j].OldestDebtDays
		})
	case "NAME_ASC":
		sort.Slice(debtors, func(i, j int) bool {
			name1 := s.getDebtorName(debtors[i])
			name2 := s.getDebtorName(debtors[j])
			return name1 < name2
		})
	default: // "DAYS_DESC" (default)
		sort.Slice(debtors, func(i, j int) bool {
			return debtors[i].OldestDebtDays > debtors[j].OldestDebtDays
		})
	}

	return debtors
}

// calculateSummary calcula resumen estadístico
func (s *reportService) calculateSummary(debtors []*input.Debtor) input.DelinquentSummary {
	summary := input.DelinquentSummary{
		TotalDebtors:         len(debtors),
		TotalDebtAmount:      0,
		IndividualDebtors:    0,
		FamilyDebtors:        0,
		AverageDaysOverdue:   0,
		AverageDebtPerDebtor: 0,
	}

	totalDays := 0

	for _, debtor := range debtors {
		summary.TotalDebtAmount += debtor.TotalDebt
		totalDays += debtor.OldestDebtDays

		if debtor.Type == debtorTypeIndividual {
			summary.IndividualDebtors++
		} else {
			summary.FamilyDebtors++
		}
	}

	if len(debtors) > 0 {
		summary.AverageDaysOverdue = totalDays / len(debtors)
		summary.AverageDebtPerDebtor = summary.TotalDebtAmount / float64(len(debtors))
	}

	return summary
}

// getDebtorName obtiene el nombre del deudor para ordenamiento
func (s *reportService) getDebtorName(debtor *input.Debtor) string {
	if debtor.Type == debtorTypeIndividual && debtor.Member != nil {
		return fmt.Sprintf("%s %s", debtor.Member.FirstName, debtor.Member.LastName)
	} else if debtor.Type == debtorTypeFamily && debtor.Family != nil {
		return debtor.Family.FamilyName
	}
	return ""
}
