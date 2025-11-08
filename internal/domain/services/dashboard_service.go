package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

type dashboardService struct {
	memberRepo   output.MemberRepository
	paymentRepo  output.PaymentRepository
	cashflowRepo output.CashFlowRepository
	familyRepo   output.FamilyRepository
	appLogger    logger.Logger
}

// Helper structures for payment processing
type paymentStats struct {
	totalPaid     float64
	monthlyPaid   float64
	lastMonthPaid float64
	pendingAmount float64
	paidCount     int
	pendingCount  int
	recentCount   int
}

type timeFilters struct {
	yearStart      time.Time
	yearEnd        time.Time
	monthStart     time.Time
	lastMonthStart time.Time
	lastMonthEnd   time.Time
	oneWeekAgo     time.Time
}

// NewDashboardService crea una nueva instancia del servicio de dashboard
func NewDashboardService(
	memberRepo output.MemberRepository,
	paymentRepo output.PaymentRepository,
	cashflowRepo output.CashFlowRepository,
	familyRepo output.FamilyRepository,
	appLogger logger.Logger,
) input.DashboardService {
	return &dashboardService{
		memberRepo:   memberRepo,
		paymentRepo:  paymentRepo,
		cashflowRepo: cashflowRepo,
		familyRepo:   familyRepo,
		appLogger:    appLogger,
	}
}

// GetDashboardStats obtiene las estadísticas principales del dashboard
func (s *dashboardService) GetDashboardStats(ctx context.Context) (*input.DashboardStats, error) {
	s.appLogger.Info("Getting dashboard stats")

	stats := &input.DashboardStats{}
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfLastMonth := startOfMonth.AddDate(0, -1, 0)
	endOfLastMonth := startOfMonth.AddDate(0, 0, -1)

	// Member statistics
	if err := s.calculateMemberStats(ctx, stats, startOfMonth, startOfLastMonth); err != nil {
		s.appLogger.Error("Error calculating member stats", zap.Error(err))
		return nil, errors.InternalError("error calculating member statistics", err)
	}

	// Payment statistics
	if err := s.calculatePaymentStats(ctx, stats, startOfMonth, startOfLastMonth, endOfLastMonth); err != nil {
		s.appLogger.Error("Error calculating payment stats", zap.Error(err))
		return nil, errors.InternalError("error calculating payment statistics", err)
	}

	// Financial statistics
	if err := s.calculateFinancialStats(ctx, stats, startOfMonth); err != nil {
		s.appLogger.Error("Error calculating financial stats", zap.Error(err))
		return nil, errors.InternalError("error calculating financial statistics", err)
	}

	// Trend data - no longer returns error
	s.calculateTrendData(ctx, stats)

	s.appLogger.Info("Dashboard stats retrieved successfully")
	return stats, nil
}

// calculateMemberStats calcula las estadísticas de miembros
func (s *dashboardService) calculateMemberStats(ctx context.Context, stats *input.DashboardStats, startOfMonth, startOfLastMonth time.Time) error {
	// Get all members - using List with empty filters to get all
	filters := output.MemberFilters{
		Page:     1,
		PageSize: 10000, // Large enough to get all members
	}

	allMembers, totalCount, err := s.memberRepo.List(ctx, filters)
	if err != nil {
		return err
	}

	stats.TotalMembers = totalCount

	// Count active, inactive, individual and family members
	for _, member := range allMembers {
		if member.State == models.EstadoActivo {
			stats.ActiveMembers++

			// Use switch for membership type comparison
			switch member.MembershipType {
			case models.TipoMembresiaPIndividual:
				stats.IndividualMembers++
			case models.TipoMembresiaPFamiliar:
				stats.FamilyMembers++
			}
		} else {
			stats.InactiveMembers++
		}

		// Count new members this month
		if member.RegistrationDate.After(startOfMonth) || member.RegistrationDate.Equal(startOfMonth) {
			stats.NewMembersThisMonth++
		}

		// Count new members last month
		if member.RegistrationDate.After(startOfLastMonth) && member.RegistrationDate.Before(startOfMonth) {
			stats.NewMembersLastMonth++
		}
	}

	// Calculate growth percentage
	if stats.NewMembersLastMonth > 0 {
		stats.MemberGrowthPercentage = float64(stats.NewMembersThisMonth-stats.NewMembersLastMonth) / float64(stats.NewMembersLastMonth) * 100
	} else if stats.NewMembersThisMonth > 0 {
		stats.MemberGrowthPercentage = 100
	}

	return nil
}

// calculatePaymentStats calcula las estadísticas de pagos
func (s *dashboardService) calculatePaymentStats(ctx context.Context, stats *input.DashboardStats, startOfMonth, startOfLastMonth, endOfLastMonth time.Time) error {
	now := time.Now()

	// Create time filters
	filters := timeFilters{
		yearStart:      time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()),
		yearEnd:        time.Date(now.Year(), 12, 31, 23, 59, 59, 0, now.Location()),
		monthStart:     startOfMonth,
		lastMonthStart: startOfLastMonth,
		lastMonthEnd:   endOfLastMonth,
		oneWeekAgo:     now.AddDate(0, 0, -7),
	}

	// Get all payments using repository filters (no duplication since all payments have member_id)
	paymentFilters := &output.PaymentRepositoryFilters{
		Limit: 10000, // Large enough to get all payments
	}

	allPayments, err := s.paymentRepo.FindAll(ctx, paymentFilters)
	if err != nil {
		return err
	}

	// Process all payments once (no more member/family separation to avoid duplication)
	paymentStats := &paymentStats{}
	s.processPayments(allPayments, paymentStats, filters)

	// Update dashboard stats
	stats.TotalRevenue = paymentStats.totalPaid
	stats.MonthlyRevenue = paymentStats.monthlyPaid
	stats.RecentPaymentsCount = paymentStats.recentCount
	stats.PendingPayments = float64(paymentStats.pendingCount)

	// Calculate average payment
	if paymentStats.paidCount > 0 {
		stats.AveragePayment = paymentStats.totalPaid / float64(paymentStats.paidCount)
	}

	// Calculate payment completion rate
	totalPayments := paymentStats.paidCount + paymentStats.pendingCount
	if totalPayments > 0 {
		stats.PaymentCompletionRate = float64(paymentStats.paidCount) / float64(totalPayments) * 100
	}

	// Calculate revenue growth percentage
	if paymentStats.lastMonthPaid > 0 {
		stats.RevenueGrowthPercentage = (paymentStats.monthlyPaid - paymentStats.lastMonthPaid) / paymentStats.lastMonthPaid * 100
	} else if paymentStats.monthlyPaid > 0 {
		stats.RevenueGrowthPercentage = 100
	}

	return nil
}


// processPayments processes a list of payments and updates statistics
func (s *dashboardService) processPayments(payments []models.Payment, stats *paymentStats, filters timeFilters) {
	for _, payment := range payments {
		// Use switch for payment status comparison
		switch payment.Status {
		case models.PaymentStatusPaid:
			stats.totalPaid += payment.Amount
			stats.paidCount++

			// This month's payments
			if payment.PaymentDate != nil && (payment.PaymentDate.After(filters.monthStart) || payment.PaymentDate.Equal(filters.monthStart)) {
				stats.monthlyPaid += payment.Amount
			}

			// Last month's payments
			if payment.PaymentDate != nil &&
				(payment.PaymentDate.After(filters.lastMonthStart) || payment.PaymentDate.Equal(filters.lastMonthStart)) &&
				(payment.PaymentDate.Before(filters.lastMonthEnd) || payment.PaymentDate.Equal(filters.lastMonthEnd)) {
				stats.lastMonthPaid += payment.Amount
			}

			// Recent payments (last week)
			if payment.PaymentDate != nil && payment.PaymentDate.After(filters.oneWeekAgo) {
				stats.recentCount++
			}

		case models.PaymentStatusPending:
			stats.pendingAmount += payment.Amount
			stats.pendingCount++

		case models.PaymentStatusCancelled:
			// Cancelled payments are ignored in statistics
			// They don't count towards revenue or pending amounts
		}
	}
}

// calculateFinancialStats calcula las estadísticas financieras
func (s *dashboardService) calculateFinancialStats(ctx context.Context, stats *input.DashboardStats, startOfMonth time.Time) error {
	// Get current balance
	balanceData, err := s.cashflowRepo.GetBalance(ctx, nil)
	if err != nil {
		return err
	}
	stats.CurrentBalance = balanceData.CurrentBalance

	// Get all transactions using List with empty filter
	filter := output.CashFlowFilter{
		Page:     1,
		PageSize: 10000, // Large enough to get all
	}

	transactions, err := s.cashflowRepo.List(ctx, filter)
	if err != nil {
		return err
	}

	stats.TotalTransactions = len(transactions)

	// Calculate monthly expenses
	var monthlyExpenses float64
	for _, transaction := range transactions {
		// Count as expense if it's an expense type and happened this month
		if transaction.OperationType.IsExpense() &&
			(transaction.Date.After(startOfMonth) || transaction.Date.Equal(startOfMonth)) {
			monthlyExpenses += transaction.Amount
		}
	}
	stats.MonthlyExpenses = monthlyExpenses

	return nil
}

// calculateTrendData calcula los datos de tendencia para los gráficos
func (s *dashboardService) calculateTrendData(ctx context.Context, stats *input.DashboardStats) {
	// Calculate membership trend for the last 6 months
	stats.MembershipTrend = s.calculateMembershipTrend(ctx, 6)

	// Calculate revenue trend for the last 6 months
	stats.RevenueTrend = s.calculateRevenueTrend(ctx, 6)
}

// calculateMembershipTrend calcula la tendencia de membresías
func (s *dashboardService) calculateMembershipTrend(ctx context.Context, months int) []input.MembershipTrendData {
	trend := make([]input.MembershipTrendData, 0, months)
	now := time.Now()

	// Get all members
	filters := output.MemberFilters{
		Page:     1,
		PageSize: 10000,
	}

	allMembers, _, err := s.memberRepo.List(ctx, filters)
	if err != nil {
		s.appLogger.Error("Error getting members for trend", zap.Error(err))
		return trend
	}

	// Calculate for each month
	for i := months - 1; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0).AddDate(0, 0, -1)
		monthName := monthStart.Format("Jan")

		var newCount, totalCount int
		for _, member := range allMembers {
			// Count members registered up to this month
			if member.RegistrationDate.Before(monthEnd) || member.RegistrationDate.Equal(monthEnd) {
				if member.State == models.EstadoActivo {
					totalCount++
				}
			}

			// Count new members in this month
			if (member.RegistrationDate.After(monthStart) || member.RegistrationDate.Equal(monthStart)) &&
				(member.RegistrationDate.Before(monthEnd) || member.RegistrationDate.Equal(monthEnd)) {
				newCount++
			}
		}

		trend = append(trend, input.MembershipTrendData{
			Month:        monthName,
			NewMembers:   newCount,
			TotalMembers: totalCount,
		})
	}

	return trend
}

// calculateRevenueTrend calcula la tendencia de ingresos y gastos
func (s *dashboardService) calculateRevenueTrend(ctx context.Context, months int) []input.RevenueTrendData {
	trend := make([]input.RevenueTrendData, 0, months)
	now := time.Now()

	// Get all transactions
	filter := output.CashFlowFilter{
		Page:     1,
		PageSize: 10000,
	}

	transactions, err := s.cashflowRepo.List(ctx, filter)
	if err != nil {
		s.appLogger.Error("Error getting transactions for trend", zap.Error(err))
		return trend
	}

	// Calculate for each month
	for i := months - 1; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0).AddDate(0, 0, -1)
		monthName := monthStart.Format("Jan")

		var revenue, expenses float64

		// Calculate revenue and expenses from transactions
		for _, transaction := range transactions {
			if (transaction.Date.After(monthStart) || transaction.Date.Equal(monthStart)) &&
				(transaction.Date.Before(monthEnd) || transaction.Date.Equal(monthEnd)) {
				if transaction.OperationType.IsIncome() {
					revenue += transaction.Amount
				} else if transaction.OperationType.IsExpense() {
					expenses += transaction.Amount
				}
			}
		}

		trend = append(trend, input.RevenueTrendData{
			Month:    monthName,
			Revenue:  revenue,
			Expenses: expenses,
		})
	}

	return trend
}

// GetRecentActivity obtiene la actividad reciente del sistema
func (s *dashboardService) GetRecentActivity(ctx context.Context, limit int) ([]*input.RecentActivity, error) {
	s.appLogger.Info("Getting recent activity", zap.Int("limit", limit))

	activities := make([]*input.RecentActivity, 0)

	// Get activities from different sources
	s.addMemberActivities(ctx, &activities)
	s.addFamilyActivities(ctx, &activities)
	s.addTransactionActivities(ctx, &activities)

	// Sort activities by timestamp (most recent first)
	s.sortActivitiesByTimestamp(activities)

	// Limit the results
	if len(activities) > limit {
		activities = activities[:limit]
	}

	s.appLogger.Info("Recent activity retrieved successfully", zap.Int("count", len(activities)))
	return activities, nil
}

// addMemberActivities adds member registration and deactivation activities
func (s *dashboardService) addMemberActivities(ctx context.Context, activities *[]*input.RecentActivity) {
	memberFilters := output.MemberFilters{
		Page:     1,
		PageSize: 100,
		OrderBy:  "registration_date DESC",
	}

	members, _, err := s.memberRepo.List(ctx, memberFilters)
	if err != nil {
		s.appLogger.Error("Error getting members for activity", zap.Error(err))
		return
	}

	for _, member := range members {
		activityID := fmt.Sprintf("member-%d", member.ID)
		activity := &input.RecentActivity{
			ID:              activityID,
			Type:            input.ActivityMemberRegistered,
			Description:     fmt.Sprintf("Nuevo miembro registrado: %s %s", member.Name, member.Surnames),
			Timestamp:       member.RegistrationDate,
			RelatedMemberID: &member.ID,
		}

		if member.State == models.EstadoInactivo && member.LeavingDate != nil {
			activity.Type = input.ActivityMemberDeactivated
			activity.Description = fmt.Sprintf("Miembro dado de baja: %s %s", member.Name, member.Surnames)
			activity.Timestamp = *member.LeavingDate
		}

		*activities = append(*activities, activity)
	}
}

// addFamilyActivities adds family creation activities
func (s *dashboardService) addFamilyActivities(ctx context.Context, activities *[]*input.RecentActivity) {
	families, _, err := s.familyRepo.List(ctx, 1, 50, nil, "created_at DESC")
	if err != nil {
		s.appLogger.Error("Error getting families for activity", zap.Error(err))
		return
	}

	for _, family := range families {
		activityID := fmt.Sprintf("family-%d", family.ID)
		description := s.buildFamilyDescription(family)

		activity := &input.RecentActivity{
			ID:              activityID,
			Type:            input.ActivityFamilyCreated,
			Description:     description,
			Timestamp:       family.CreatedAt,
			RelatedFamilyID: &family.ID,
		}
		*activities = append(*activities, activity)
	}
}

// buildFamilyDescription generates a descriptive name for a family using spouse names
func (s *dashboardService) buildFamilyDescription(family *models.Family) string {
	if family.EsposoNombre == "" && family.EsposaNombre == "" {
		return fmt.Sprintf("Nueva familia: %s", family.NumeroSocio)
	}

	names := make([]string, 0, 2)
	if family.EsposoNombre != "" {
		names = append(names, family.EsposoNombre)
	}
	if family.EsposaNombre != "" {
		names = append(names, family.EsposaNombre)
	}

	if len(names) > 0 {
		return fmt.Sprintf("Nueva familia: %s", strings.Join(names, " y "))
	}

	return fmt.Sprintf("Nueva familia: %s", family.NumeroSocio)
}

// addTransactionActivities adds cash flow transaction activities
func (s *dashboardService) addTransactionActivities(ctx context.Context, activities *[]*input.RecentActivity) {
	transactionFilter := output.CashFlowFilter{
		Page:     1,
		PageSize: 50,
		OrderBy:  "date DESC",
	}

	transactions, err := s.cashflowRepo.List(ctx, transactionFilter)
	if err != nil {
		s.appLogger.Error("Error getting transactions for activity", zap.Error(err))
		return
	}

	for _, transaction := range transactions {
		activityID := fmt.Sprintf("transaction-%d", transaction.ID)
		activity := &input.RecentActivity{
			ID:          activityID,
			Type:        input.ActivityTransactionRecorded,
			Description: fmt.Sprintf("Transacción registrada: %s - €%.2f", transaction.Detail, transaction.Amount),
			Timestamp:   transaction.Date,
			Amount:      &transaction.Amount,
		}

		if transaction.MemberID != nil {
			activity.RelatedMemberID = transaction.MemberID
		}

		*activities = append(*activities, activity)
	}
}

// sortActivitiesByTimestamp sorts activities by timestamp in descending order
func (s *dashboardService) sortActivitiesByTimestamp(activities []*input.RecentActivity) {
	// Simple bubble sort for now - can optimize if needed
	n := len(activities)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if activities[j].Timestamp.Before(activities[j+1].Timestamp) {
				activities[j], activities[j+1] = activities[j+1], activities[j]
			}
		}
	}
}
