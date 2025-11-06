package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

const (
	debtorTypeIndividual = "INDIVIDUAL"
	debtorTypeFamily     = "FAMILY"
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
	// 1. Establecer fecha de corte (default: hoy)
	cutoffDate := time.Now()
	if inputParams.CutoffDate != nil {
		cutoffDate = *inputParams.CutoffDate
	}

	// 2. Obtener todos los pagos PENDIENTES
	pendingPayments, err := s.reportRepo.GetPendingPayments(ctx)
	if err != nil {
		return nil, errors.DB(err, "error obteniendo pagos pendientes")
	}

	// 3. Agrupar pagos por deudor (member_id o family_id)
	debtorMap := make(map[string]*input.Debtor)

	for i := range pendingPayments {
		payment := &pendingPayments[i]

		// Determinar el tipo de deudor
		var debtorKey string
		var debtorType string
		var entityID uint

		if payment.MemberID != 0 {
			debtorKey = fmt.Sprintf("member_%d", payment.MemberID)
			debtorType = debtorTypeIndividual
			entityID = payment.MemberID
		} else {
			// Skip pagos sin member_id (datos inconsistentes)
			continue
		}

		// Si el deudor no existe en el map, crearlo
		if _, exists := debtorMap[debtorKey]; !exists {
			debtor := &input.Debtor{
				Type:            debtorType,
				PendingPayments: []*input.PendingPayment{},
				TotalDebt:       0,
			}

			// Cargar información del socio o familia
			if debtorType == debtorTypeIndividual {
				memberInfo, err := s.loadMemberInfo(ctx, entityID)
				if err != nil {
					// Log y continuar con el siguiente deudor
					continue
				}
				debtor.MemberID = &entityID
				debtor.Member = memberInfo

				// Obtener último pago PAID de este socio
				lastPaid, err := s.reportRepo.GetLastPaidPaymentForMember(ctx, entityID)
				if err != nil {
					// Log y continuar
				} else if lastPaid != nil {
					debtor.LastPaymentDate = &lastPaid.PaymentDate
					debtor.LastPaymentAmount = &lastPaid.Amount
				}
			}

			debtorMap[debtorKey] = debtor
		}

		// Añadir el pago pendiente al deudor
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

		debtorMap[debtorKey].PendingPayments = append(
			debtorMap[debtorKey].PendingPayments,
			pendingPayment,
		)

		debtorMap[debtorKey].TotalDebt += payment.Amount

		// Actualizar el pago más antiguo
		if debtorMap[debtorKey].OldestDebtDate.IsZero() ||
			payment.CreatedAt.Before(debtorMap[debtorKey].OldestDebtDate) {
			debtorMap[debtorKey].OldestDebtDate = payment.CreatedAt
			debtorMap[debtorKey].OldestDebtDays = daysOverdue
		}
	}

	// 4. Convertir map a slice
	debtors := make([]*input.Debtor, 0, len(debtorMap))
	for _, debtor := range debtorMap {
		debtors = append(debtors, debtor)
	}

	// 5. Aplicar filtros
	debtors = s.applyFilters(debtors, inputParams)

	// 6. Ordenar según inputParams.SortBy
	debtors = s.sortDebtors(debtors, inputParams.SortBy)

	// 7. Calcular resumen estadístico
	summary := s.calculateSummary(debtors)

	// 8. Retornar respuesta
	return &input.DelinquentReportResponse{
		Debtors:     debtors,
		Summary:     summary,
		GeneratedAt: time.Now(),
	}, nil
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

	// Nota: El modelo Member no tiene campo Phone directo,
	// podría estar en observaciones o en otro lugar según el schema
	// Por ahora lo dejamos nil

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
