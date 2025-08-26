package resolvers

import (
	"context"
	"fmt"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
)

// dashboardResolver es una estructura interna que contiene el resolver principal
type dashboardResolver struct {
	*Resolver
}

// GetDashboardStats implementa el resolver para obtener las estadísticas del dashboard
func (r *dashboardResolver) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	// Get dashboard stats from service
	stats, err := r.dashboardService.GetDashboardStats(ctx)
	if err != nil {
		return nil, err
	}

	// Map domain stats to GraphQL model
	gqlStats := &model.DashboardStats{
		// Member stats
		TotalMembers:           stats.TotalMembers,
		ActiveMembers:          stats.ActiveMembers,
		InactiveMembers:        stats.InactiveMembers,
		IndividualMembers:      stats.IndividualMembers,
		FamilyMembers:          stats.FamilyMembers,
		NewMembersThisMonth:    stats.NewMembersThisMonth,
		NewMembersLastMonth:    stats.NewMembersLastMonth,
		MemberGrowthPercentage: stats.MemberGrowthPercentage,

		// Payment stats
		TotalRevenue:            stats.TotalRevenue,
		MonthlyRevenue:          stats.MonthlyRevenue,
		PendingPayments:         stats.PendingPayments,
		AveragePayment:          stats.AveragePayment,
		PaymentCompletionRate:   stats.PaymentCompletionRate,
		RevenueGrowthPercentage: stats.RevenueGrowthPercentage,

		// Financial stats
		CurrentBalance:  stats.CurrentBalance,
		MonthlyExpenses: stats.MonthlyExpenses,

		// Activity stats
		TotalTransactions:   stats.TotalTransactions,
		RecentPaymentsCount: stats.RecentPaymentsCount,

		// Time-based data for charts
		MembershipTrend: mapMembershipTrend(stats.MembershipTrend),
		RevenueTrend:    mapRevenueTrend(stats.RevenueTrend),
	}

	return gqlStats, nil
}

// GetRecentActivity implementa el resolver para obtener la actividad reciente
func (r *dashboardResolver) GetRecentActivity(ctx context.Context, limit *int) ([]*model.RecentActivity, error) {
	// Default limit
	activityLimit := 10
	if limit != nil && *limit > 0 {
		activityLimit = *limit
		// Cap at 50 for performance
		if activityLimit > 50 {
			activityLimit = 50
		}
	}

	// Get recent activity from service
	activities, err := r.dashboardService.GetRecentActivity(ctx, activityLimit)
	if err != nil {
		return nil, err
	}

	// Map domain activities to GraphQL model
	gqlActivities := make([]*model.RecentActivity, 0, len(activities))
	for _, activity := range activities {
		gqlActivity := &model.RecentActivity{
			ID:          uintToString(activity.ID),
			Type:        mapActivityType(activity.Type),
			Description: activity.Description,
			Timestamp:   activity.Timestamp,
			Amount:      activity.Amount,
		}

		// Load related entities if needed
		if activity.RelatedMemberID != nil {
			member, err := r.memberService.GetMemberByID(ctx, *activity.RelatedMemberID)
			if err == nil && member != nil {
				gqlActivity.RelatedMember = mapMemberToGraphQL(member)
			}
		}

		if activity.RelatedFamilyID != nil {
			family, err := r.familyService.GetByID(ctx, *activity.RelatedFamilyID)
			if err == nil && family != nil {
				gqlActivity.RelatedFamily = mapFamilyToGraphQL(family)
			}
		}

		gqlActivities = append(gqlActivities, gqlActivity)
	}

	return gqlActivities, nil
}

// Helper functions for mapping

func mapActivityType(activityType input.ActivityType) model.ActivityType {
	switch activityType {
	case input.ActivityMemberRegistered:
		return model.ActivityTypeMemberRegistered
	case input.ActivityPaymentReceived:
		return model.ActivityTypePaymentReceived
	case input.ActivityFamilyCreated:
		return model.ActivityTypeFamilyCreated
	case input.ActivityMemberDeactivated:
		return model.ActivityTypeMemberDeactivated
	case input.ActivityTransactionRecorded:
		return model.ActivityTypeTransactionRecorded
	default:
		return model.ActivityTypeTransactionRecorded
	}
}

func mapMembershipTrend(trends []input.MembershipTrendData) []*model.MembershipTrendData {
	gqlTrends := make([]*model.MembershipTrendData, 0, len(trends))
	for _, trend := range trends {
		gqlTrends = append(gqlTrends, &model.MembershipTrendData{
			Month:        trend.Month,
			NewMembers:   trend.NewMembers,
			TotalMembers: trend.TotalMembers,
		})
	}
	return gqlTrends
}

func mapRevenueTrend(trends []input.RevenueTrendData) []*model.RevenueTrendData {
	gqlTrends := make([]*model.RevenueTrendData, 0, len(trends))
	for _, trend := range trends {
		gqlTrends = append(gqlTrends, &model.RevenueTrendData{
			Month:    trend.Month,
			Revenue:  trend.Revenue,
			Expenses: trend.Expenses,
		})
	}
	return gqlTrends
}

func uintToString(id uint) string {
	return fmt.Sprintf("%d", id)
}

// mapMemberToGraphQL convierte un modelo de dominio Member a GraphQL Member
func mapMemberToGraphQL(member *models.Member) *models.Member {
	if member == nil {
		return nil
	}

	// Retornamos directamente el member del dominio ya que
	// el resolver Member se encarga de mapear los campos correctamente
	return member
}

// mapFamilyToGraphQL convierte un modelo de dominio Family a GraphQL Family
func mapFamilyToGraphQL(family *models.Family) *models.Family {
	if family == nil {
		return nil
	}

	// Retornamos directamente el family del dominio ya que
	// el resolver Family se encarga de mapear los campos correctamente
	return family
}
