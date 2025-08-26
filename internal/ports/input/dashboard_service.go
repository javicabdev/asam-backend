package input

import (
	"context"
	"time"
)

// DashboardService define las operaciones de negocio para el dashboard
type DashboardService interface {
	// GetDashboardStats obtiene las estadísticas principales del dashboard
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)

	// GetRecentActivity obtiene la actividad reciente del sistema
	GetRecentActivity(ctx context.Context, limit int) ([]*RecentActivity, error)
}

// DashboardStats contiene las estadísticas principales del dashboard
type DashboardStats struct {
	// Member stats
	TotalMembers           int
	ActiveMembers          int
	InactiveMembers        int
	IndividualMembers      int
	FamilyMembers          int
	NewMembersThisMonth    int
	NewMembersLastMonth    int
	MemberGrowthPercentage float64

	// Payment stats
	TotalRevenue            float64
	MonthlyRevenue          float64
	PendingPayments         float64
	AveragePayment          float64
	PaymentCompletionRate   float64
	RevenueGrowthPercentage float64

	// Financial stats
	CurrentBalance  float64
	MonthlyExpenses float64

	// Activity stats
	TotalTransactions   int
	RecentPaymentsCount int

	// Time-based data for charts
	MembershipTrend []MembershipTrendData
	RevenueTrend    []RevenueTrendData
}

// MembershipTrendData contiene datos de tendencia de membresías
type MembershipTrendData struct {
	Month        string
	NewMembers   int
	TotalMembers int
}

// RevenueTrendData contiene datos de tendencia de ingresos
type RevenueTrendData struct {
	Month    string
	Revenue  float64
	Expenses float64
}

// ActivityType representa el tipo de actividad
type ActivityType string

const (
	ActivityMemberRegistered    ActivityType = "MEMBER_REGISTERED"
	ActivityPaymentReceived     ActivityType = "PAYMENT_RECEIVED"
	ActivityFamilyCreated       ActivityType = "FAMILY_CREATED"
	ActivityMemberDeactivated   ActivityType = "MEMBER_DEACTIVATED"
	ActivityTransactionRecorded ActivityType = "TRANSACTION_RECORDED"
)

// RecentActivity representa una actividad reciente en el sistema
type RecentActivity struct {
	ID              uint
	Type            ActivityType
	Description     string
	Timestamp       time.Time
	RelatedMemberID *uint
	RelatedFamilyID *uint
	Amount          *float64
}
