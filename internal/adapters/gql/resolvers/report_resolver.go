package resolvers

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// reportResolver es una estructura interna que contiene el resolver principal
type reportResolver struct {
	*Resolver
}

// GetDelinquentReport implementa el resolver para obtener el reporte de morosos
func (r *reportResolver) GetDelinquentReport(ctx context.Context, inputParams *model.DelinquentReportInput) (*model.DelinquentReportResponse, error) {
	// Obtener usuario del contexto
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized, "unauthorized: authentication required")
	}

	// SOLO ADMIN puede acceder a este informe
	if !user.IsAdmin() {
		return nil, errors.New(errors.ErrForbidden, "forbidden: only admin can access delinquent report")
	}

	// Convertir input GraphQL a input del servicio
	serviceInput := input.DelinquentReportInput{
		Page:     1,
		PageSize: 10,
	}

	if inputParams != nil {
		serviceInput.CutoffDate = inputParams.CutoffDate
		serviceInput.MinAmount = inputParams.MinAmount
		serviceInput.DebtorType = inputParams.DebtorType
		serviceInput.SortBy = inputParams.SortBy

		// Extraer paginación
		if inputParams.Pagination != nil {
			serviceInput.Page = inputParams.Pagination.Page
			serviceInput.PageSize = inputParams.Pagination.PageSize
		}
	}

	// Llamar al servicio
	report, err := r.reportService.GetDelinquentReport(ctx, serviceInput)
	if err != nil {
		return nil, err
	}

	// Convertir respuesta del servicio a GraphQL con pageInfo
	return mapDelinquentReportToGraphQL(report, serviceInput.Page, serviceInput.PageSize), nil
}

// mapDelinquentReportToGraphQL convierte la respuesta del servicio a modelo GraphQL
func mapDelinquentReportToGraphQL(report *input.DelinquentReportResponse, page int, pageSize int) *model.DelinquentReportResponse {
	debtors := make([]*model.Debtor, len(report.Debtors))
	for i, debtor := range report.Debtors {
		debtors[i] = mapDebtorToGraphQL(debtor)
	}

	// Calcular paginación
	totalPages := (report.TotalCount + pageSize - 1) / pageSize
	hasNextPage := page < totalPages
	hasPreviousPage := page > 1

	pageInfo := &model.PageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: hasPreviousPage,
		TotalCount:      report.TotalCount,
	}

	return &model.DelinquentReportResponse{
		Debtors:     debtors,
		PageInfo:    pageInfo,
		Summary:     mapSummaryToGraphQL(&report.Summary),
		GeneratedAt: report.GeneratedAt,
	}
}

// mapDebtorToGraphQL convierte un Debtor del servicio a GraphQL
func mapDebtorToGraphQL(debtor *input.Debtor) *model.Debtor {
	gqlDebtor := &model.Debtor{
		Type:            debtor.Type,
		PendingPayments: mapPendingPaymentsToGraphQL(debtor.PendingPayments),
		TotalDebt:       debtor.TotalDebt,
		OldestDebtDays:  debtor.OldestDebtDays,
		OldestDebtDate:  debtor.OldestDebtDate,
	}

	if debtor.MemberID != nil {
		memberID := uintToString(*debtor.MemberID)
		gqlDebtor.MemberID = &memberID
	}

	if debtor.FamilyID != nil {
		familyID := uintToString(*debtor.FamilyID)
		gqlDebtor.FamilyID = &familyID
	}

	if debtor.Member != nil {
		gqlDebtor.Member = mapDebtorMemberInfoToGraphQL(debtor.Member)
	}

	if debtor.Family != nil {
		gqlDebtor.Family = mapDebtorFamilyInfoToGraphQL(debtor.Family)
	}

	gqlDebtor.LastPaymentDate = debtor.LastPaymentDate
	gqlDebtor.LastPaymentAmount = debtor.LastPaymentAmount

	return gqlDebtor
}

// mapDebtorMemberInfoToGraphQL convierte información de socio a GraphQL
func mapDebtorMemberInfoToGraphQL(member *input.DebtorMemberInfo) *model.DebtorMemberInfo {
	return &model.DebtorMemberInfo{
		ID:           uintToString(member.ID),
		MemberNumber: member.MemberNumber,
		FirstName:    member.FirstName,
		LastName:     member.LastName,
		Email:        member.Email,
		Phone:        member.Phone,
		Status:       member.Status,
	}
}

// mapDebtorFamilyInfoToGraphQL convierte información de familia a GraphQL
func mapDebtorFamilyInfoToGraphQL(family *input.DebtorFamilyInfo) *model.DebtorFamilyInfo {
	return &model.DebtorFamilyInfo{
		ID:            uintToString(family.ID),
		FamilyName:    family.FamilyName,
		PrimaryMember: mapDebtorMemberInfoToGraphQL(&family.PrimaryMember),
		TotalMembers:  family.TotalMembers,
	}
}

// mapPendingPaymentsToGraphQL convierte lista de pagos pendientes a GraphQL
func mapPendingPaymentsToGraphQL(payments []*input.PendingPayment) []*model.PendingPayment {
	gqlPayments := make([]*model.PendingPayment, len(payments))
	for i, payment := range payments {
		gqlPayment := &model.PendingPayment{
			ID:          uintToString(payment.ID),
			Amount:      payment.Amount,
			CreatedAt:   payment.CreatedAt,
			DaysOverdue: payment.DaysOverdue,
			Notes:       payment.Notes,
		}
		gqlPayments[i] = gqlPayment
	}
	return gqlPayments
}

// mapSummaryToGraphQL convierte el resumen a GraphQL
func mapSummaryToGraphQL(summary *input.DelinquentSummary) *model.DelinquentSummary {
	return &model.DelinquentSummary{
		TotalDebtors:         summary.TotalDebtors,
		IndividualDebtors:    summary.IndividualDebtors,
		FamilyDebtors:        summary.FamilyDebtors,
		TotalDebtAmount:      summary.TotalDebtAmount,
		AverageDaysOverdue:   summary.AverageDaysOverdue,
		AverageDebtPerDebtor: summary.AverageDebtPerDebtor,
	}
}
