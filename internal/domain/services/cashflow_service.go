package services

import (
	"context"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/metrics"
)

// CashFlowService implementa input.CashFlowService
type CashFlowService struct {
	repository output.CashFlowRepository
}

// NewCashFlowService crea una nueva instancia de CashFlowService
func NewCashFlowService(repository output.CashFlowRepository) *CashFlowService {
	return &CashFlowService{
		repository: repository,
	}
}

// RegisterMovement implementa la creación de un nuevo movimiento
func (s *CashFlowService) RegisterMovement(ctx context.Context, movement *models.CashFlow) error {
	// Validar el movimiento
	if err := movement.Validate(); err != nil {
		return errors.NewValidationError(err.Error(), nil)
	}

	// Registrar el movimiento
	if err := s.repository.Create(ctx, movement); err != nil {
		return errors.DB(err, "error registrando movimiento")
	}

	// Actualizar métricas de flujo de caja
	metrics.CashFlowMetrics.WithLabelValues(
		string(movement.OperationType),
	).Add(movement.Amount)

	// Si es un gasto, el amount ya viene negativo
	if movement.OperationType.IsIncome() {
		metrics.PaymentMetrics.WithLabelValues("income", "completed").Add(movement.Amount)
	} else {
		metrics.PaymentMetrics.WithLabelValues("expense", "completed").Add(math.Abs(movement.Amount))
	}

	return nil
}

// GetMovement implementa la obtención de un movimiento por ID
func (s *CashFlowService) GetMovement(ctx context.Context, id uint) (*models.CashFlow, error) {
	movement, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, errors.DB(err, "error obteniendo movimiento por ID")
	}

	if movement == nil {
		return nil, errors.NotFound("movement", nil)
	}

	return movement, nil
}

// GetMovementsByPeriod obtiene los movimientos de caja en un período específico
func (s *CashFlowService) GetMovementsByPeriod(ctx context.Context, filter input.CashFlowFilter) ([]*models.CashFlow, error) {
	// Validaciones básicas
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	if filter.Page < 1 {
		filter.Page = 1
	}

	// Validar ordenamiento si se proporciona
	if filter.OrderBy != "" {
		// Validar que el campo de ordenamiento sea válido
		validFields := map[string]bool{
			"date":           true,
			"amount":         true,
			"operation_type": true,
		}

		// Extraer el campo de ordenamiento (quitando el ASC/DESC)
		parts := strings.Fields(filter.OrderBy)
		if !validFields[strings.ToLower(parts[0])] {
			return nil, errors.Validation("Campo de ordenamiento inválido", "orderBy", parts[0])
		}
	}

	// Convertir el filtro de input a output
	repoFilter := output.CashFlowFilter{
		MemberID:      filter.MemberID,
		StartDate:     filter.StartDate,
		EndDate:       filter.EndDate,
		OperationType: filter.OperationType,
		Page:          filter.Page,
		PageSize:      filter.PageSize,
		OrderBy:       filter.OrderBy,
	}

	// Obtener los movimientos usando el repositorio
	movements, err := s.repository.List(ctx, repoFilter)
	if err != nil {
		return nil, errors.DB(err, "error obteniendo movimientos del período")
	}

	return movements, nil
}

// UpdateMovement implementa la actualización de un movimiento
func (s *CashFlowService) UpdateMovement(ctx context.Context, movement *models.CashFlow) error {
	// Validar el movimiento
	if err := movement.Validate(); err != nil {
		return errors.Validation("Error validating movement", "", err.Error())
	}

	if err := s.repository.Update(ctx, movement); err != nil {
		return errors.DB(err, "error actualizando movimiento")
	}

	return nil
}

// DeleteMovement implementa el borrado de un movimiento
func (s *CashFlowService) DeleteMovement(ctx context.Context, id uint) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		return errors.DB(err, "error eliminando movimiento")
	}
	return nil
}

// GetCurrentBalance obtiene el balance actual con detalles
func (s *CashFlowService) GetCurrentBalance(ctx context.Context) (*input.BalanceReport, error) {
	// Obtener el balance actual
	balanceData, err := s.repository.GetBalance(ctx, nil)
	if err != nil {
		return nil, errors.DB(err, "error al obtener balance")
	}
	currentBalance := balanceData.CurrentBalance

	// Obtener movimientos del período actual (mes en curso)
	startOfMonth := time.Now().UTC().AddDate(0, 0, -time.Now().UTC().Day()+1)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	filter := output.CashFlowFilter{
		StartDate: &startOfMonth,
		EndDate:   &endOfMonth,
	}

	movements, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.DB(err, "error al obtener movimientos")
	}

	// Calcular totales por tipo de operación
	var totalIncome, totalExpenses float64
	summaryMap := make(map[models.OperationType]*input.MovementSummary)

	for _, mov := range movements {
		if mov.OperationType.IsIncome() {
			totalIncome += mov.Amount
		} else {
			totalExpenses += mov.Amount
		}

		// Actualizar resumen por tipo de operación
		if summary, exists := summaryMap[mov.OperationType]; exists {
			summary.Total += mov.Amount
			summary.Count++
		} else {
			summaryMap[mov.OperationType] = &input.MovementSummary{
				OperationType: mov.OperationType,
				Total:         mov.Amount,
				Count:         1,
			}
		}
	}

	// Convertir map a slice
	summaries := make([]input.MovementSummary, 0, len(summaryMap))
	for _, summary := range summaryMap {
		summaries = append(summaries, *summary)
	}

	// Actualizar métrica de balance actual
	metrics.CashFlowMetrics.WithLabelValues("balance").Set(currentBalance)

	return &input.BalanceReport{
		CurrentBalance:   currentBalance,
		TotalIncome:      totalIncome,
		TotalExpenses:    totalExpenses,
		PeriodStart:      startOfMonth,
		PeriodEnd:        endOfMonth,
		MovementsSummary: summaries,
	}, nil
}

// GetBalanceByPeriod obtiene el balance para un período específico
func (s *CashFlowService) GetBalanceByPeriod(ctx context.Context, startDate, endDate time.Time) (*input.BalanceReport, error) {
	filter := output.CashFlowFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	movements, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.DB(err, "error al obtener movimientos")
	}

	var currentBalance, totalIncome, totalExpenses float64
	summaryMap := make(map[models.OperationType]*input.MovementSummary)

	for _, mov := range movements {
		currentBalance += mov.Amount

		if mov.OperationType.IsIncome() {
			totalIncome += mov.Amount
		} else {
			totalExpenses += mov.Amount
		}

		if summary, exists := summaryMap[mov.OperationType]; exists {
			summary.Total += mov.Amount
			summary.Count++
		} else {
			summaryMap[mov.OperationType] = &input.MovementSummary{
				OperationType: mov.OperationType,
				Total:         mov.Amount,
				Count:         1,
			}
		}
	}

	var summaries = make([]input.MovementSummary, 0, len(summaryMap))
	for _, summary := range summaryMap {
		summaries = append(summaries, *summary)
	}

	return &input.BalanceReport{
		CurrentBalance:   currentBalance,
		TotalIncome:      totalIncome,
		TotalExpenses:    totalExpenses,
		PeriodStart:      startDate,
		PeriodEnd:        endDate,
		MovementsSummary: summaries,
	}, nil
}

// ValidateBalance implementa la validación del balance actual
func (s *CashFlowService) ValidateBalance(ctx context.Context) (*input.BalanceValidation, error) {
	// Obtener todos los movimientos
	filter := output.CashFlowFilter{
		PageSize: 0, // Sin límite para obtener todos
	}

	movements, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.DB(err, "error al obtener movimientos")
	}

	// Calcular balance esperado
	var expectedBalance float64
	for _, mov := range movements {
		expectedBalance += mov.Amount
	}

	// Obtener balance actual
	balanceData, err := s.repository.GetBalance(ctx, nil)
	if err != nil {
		return nil, errors.DB(err, "error al obtener balance actual")
	}
	actualBalance := balanceData.CurrentBalance

	// Calcular discrepancia
	discrepancy := actualBalance - expectedBalance

	return &input.BalanceValidation{
		IsValid:         discrepancy == 0,
		ExpectedBalance: expectedBalance,
		ActualBalance:   actualBalance,
		Discrepancy:     discrepancy,
		Details:         s.getValidationDetails(discrepancy),
	}, nil
}

// getValidationDetails genera detalles para la validación del balance
func (s *CashFlowService) getValidationDetails(discrepancy float64) string {
	if discrepancy == 0 {
		return "El balance está correcto"
	}
	if discrepancy > 0 {
		return "El balance actual excede al esperado por " + formatAmount(discrepancy)
	}
	return "El balance actual es menor al esperado por " + formatAmount(-discrepancy)
}

// formatAmount formatea un importe para presentación
func formatAmount(amount float64) string {
	return strings.TrimRight(strings.TrimRight(formatFloat(amount, 2), "0"), ".")
}

// formatFloat formatea un número flotante con precisión específica
func formatFloat(num float64, precision int) string {
	format := "%." + string(rune(precision+'0')) + "f"
	return sprintf(format, num)
}

// sprintf es un wrapper para fmt.Sprintf
func sprintf(format string, args ...any) string {
	// Implementación simplificada para evitar importar fmt
	result := format
	for _, arg := range args {
		result += toString(arg)
	}
	return result
}

// toString convierte un valor a string
func toString(v any) string {
	switch val := v.(type) {
	case float64:
		return strconv.FormatFloat(val, 'f', 2, 64)
	default:
		return ""
	}
}

// GetFinancialReport genera reportes financieros según el tipo solicitado
func (s *CashFlowService) GetFinancialReport(ctx context.Context, reportType input.ReportType, period input.Period) (*input.FinancialReport, error) {
	filter := output.CashFlowFilter{
		StartDate: &period.StartDate,
		EndDate:   &period.EndDate,
	}

	movements, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.DB(err, "error al obtener movimientos")
	}

	// Procesar movimientos según el tipo de reporte
	switch reportType {
	case input.ReportTypeBalance:
		return s.generateBalanceReport(movements, period)
	case input.ReportTypeIncome:
		return s.generateIncomeReport(movements, period)
	case input.ReportTypeCashFlow:
		return s.generateCashFlowReport(movements, period)
	default:
		return nil, errors.New(errors.ErrInvalidFormat, "tipo de reporte no soportado: "+string(reportType))
	}
}

// generateBalanceReport genera un reporte de balance
func (s *CashFlowService) generateBalanceReport(movements []*models.CashFlow, period input.Period) (*input.FinancialReport, error) {
	report := &input.FinancialReport{
		Type:       input.ReportTypeBalance,
		Period:     period,
		Data:       make(map[string]float64),
		Categories: []input.CategorySummary{},
	}

	categorySums := make(map[string]float64)
	var totalIncome, totalExpenses float64

	for _, mov := range movements {
		category := string(mov.OperationType)
		categorySums[category] += mov.Amount

		if mov.OperationType.IsIncome() {
			totalIncome += mov.Amount
		} else {
			totalExpenses += mov.Amount
		}
	}

	// Calcular porcentajes y crear resúmenes por categoría
	total := totalIncome + totalExpenses
	for category, amount := range categorySums {
		percentage := 0.0
		if total != 0 {
			percentage = (amount / total) * 100
		}

		report.Categories = append(report.Categories, input.CategorySummary{
			Category:   category,
			Amount:     amount,
			Percentage: percentage,
		})
	}

	report.Totals = input.TotalsSummary{
		Income:    totalIncome,
		Expenses:  totalExpenses,
		NetResult: totalIncome + totalExpenses,
	}

	return report, nil
}

// GetCashFlowTrends analiza las tendencias de flujo de caja
func (s *CashFlowService) GetCashFlowTrends(ctx context.Context, period input.Period) (*input.TrendAnalysis, error) {
	// Obtener movimientos del período
	filter := output.CashFlowFilter{
		StartDate: &period.StartDate,
		EndDate:   &period.EndDate,
	}

	movements, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.DB(err, "error al obtener movimientos")
	}

	// Agrupar movimientos por mes
	monthlyData := make(map[string]*monthlyStats)
	for _, mov := range movements {
		monthKey := mov.Date.Format("2006-01")
		if _, exists := monthlyData[monthKey]; !exists {
			monthlyData[monthKey] = &monthlyStats{
				month: time.Date(mov.Date.Year(), mov.Date.Month(), 1, 0, 0, 0, 0, time.UTC),
			}
		}

		stats := monthlyData[monthKey]
		if mov.OperationType.IsIncome() {
			stats.income += mov.Amount
		} else {
			stats.expenses += mov.Amount
		}
		stats.balance += mov.Amount
	}

	// Calcular tendencias mensuales
	trends := make([]input.MonthlyTrend, 0, len(monthlyData))
	var prevBalance float64

	// Ordenar meses
	months := make([]string, 0, len(monthlyData))
	for month := range monthlyData {
		months = append(months, month)
	}
	sort.Strings(months)

	for _, month := range months {
		stats := monthlyData[month]
		growth := 0.0
		if prevBalance != 0 {
			growth = ((stats.balance - prevBalance) / math.Abs(prevBalance)) * 100
		}

		trends = append(trends, input.MonthlyTrend{
			Month:    stats.month,
			Income:   stats.income,
			Expenses: stats.expenses,
			Balance:  stats.balance,
			Growth:   growth,
		})

		prevBalance = stats.balance
	}

	// Calcular indicadores
	indicators := calculateIndicators(trends)

	return &input.TrendAnalysis{
		Period:        period,
		MonthlyTrends: trends,
		Indicators:    indicators,
	}, nil
}

// monthlyStats es una estructura auxiliar para agrupar estadísticas mensuales
type monthlyStats struct {
	month    time.Time
	income   float64
	expenses float64
	balance  float64
}

// calculateIndicators calcula los indicadores financieros basados en las tendencias
func calculateIndicators(trends []input.MonthlyTrend) map[string]float64 {
	if len(trends) == 0 {
		return make(map[string]float64)
	}

	indicators := make(map[string]float64)

	// Crecimiento promedio
	var totalGrowth float64
	for _, trend := range trends {
		totalGrowth += trend.Growth
	}
	indicators["avgGrowth"] = totalGrowth / float64(len(trends))

	// Ratio ingreso/gasto promedio
	var totalRatio float64
	ratioCount := 0
	for _, trend := range trends {
		if trend.Expenses != 0 {
			totalRatio += math.Abs(trend.Income / trend.Expenses)
			ratioCount++
		}
	}
	if ratioCount > 0 {
		indicators["avgIncomeExpenseRatio"] = totalRatio / float64(ratioCount)
	}

	return indicators
}

// GetProjections genera proyecciones financieras
func (s *CashFlowService) GetProjections(ctx context.Context, months int) (*input.FinancialProjection, error) {
	// Obtener datos históricos de los últimos 12 meses
	endDate := time.Now()
	startDate := endDate.AddDate(0, -12, 0)

	filter := output.CashFlowFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	movements, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.DB(err, "error al obtener datos históricos")
	}

	// Calcular promedios mensuales
	monthlyStats := calculateMonthlyAverages(movements)

	// Generar proyecciones
	projections := make([]input.MonthlyProjection, months)
	baseMonth := endDate

	for i := 0; i < months; i++ {
		projMonth := baseMonth.AddDate(0, i+1, 0)
		variance := calculateVariance(monthlyStats, float64(i+1))

		projections[i] = input.MonthlyProjection{
			Month:            projMonth,
			ExpectedIncome:   monthlyStats.avgIncome * (1 + variance),
			ExpectedExpenses: monthlyStats.avgExpenses * (1 + variance),
			ExpectedBalance:  monthlyStats.avgBalance * (1 + variance),
			Variance:         variance,
		}
	}

	return &input.FinancialProjection{
		Months:      months,
		Projections: projections,
		Confidence:  calculateConfidence(monthlyStats.stdDev, months),
	}, nil
}

// calculateMonthlyAverages calcula estadísticas mensuales para proyecciones
type monthlyAverages struct {
	avgIncome   float64
	avgExpenses float64
	avgBalance  float64
	stdDev      float64
}

// calculateMonthlyAverages calcula estadísticas mensuales para proyecciones
func calculateMonthlyAverages(movements []*models.CashFlow) monthlyAverages {
	var stats monthlyAverages
	if len(movements) == 0 {
		return stats
	}

	// Agrupar por mes
	monthlyData := make(map[string]*monthlyStats)
	for _, mov := range movements {
		monthKey := mov.Date.Format("2006-01")
		if _, exists := monthlyData[monthKey]; !exists {
			monthlyData[monthKey] = &monthlyStats{}
		}

		if mov.OperationType.IsIncome() {
			monthlyData[monthKey].income += mov.Amount
		} else {
			monthlyData[monthKey].expenses += mov.Amount
		}
		monthlyData[monthKey].balance += mov.Amount
	}

	// Calcular promedios
	numMonths := float64(len(monthlyData))
	var balances = make([]float64, 0, len(monthlyData))

	for _, monthly := range monthlyData {
		stats.avgIncome += monthly.income
		stats.avgExpenses += monthly.expenses
		stats.avgBalance += monthly.balance
		balances = append(balances, monthly.balance)
	}

	stats.avgIncome /= numMonths
	stats.avgExpenses /= numMonths
	stats.avgBalance /= numMonths

	// Calcular desviación estándar del balance
	var sumSquares float64
	for _, balance := range balances {
		diff := balance - stats.avgBalance
		sumSquares += diff * diff
	}
	stats.stdDev = math.Sqrt(sumSquares / numMonths)

	return stats
}

// calculateVariance calcula la varianza esperada basada en estadísticas históricas
func calculateVariance(stats monthlyAverages, monthsAhead float64) float64 {
	// Base variance starts at 2%
	baseVariance := 0.02

	// Increase variance based on historical volatility (stdDev)
	if stats.avgBalance != 0 {
		volatility := stats.stdDev / math.Abs(stats.avgBalance)
		baseVariance += volatility
	}

	// Increase variance with time horizon
	return baseVariance * monthsAhead
}

// calculateConfidence calcula el nivel de confianza de las proyecciones
func calculateConfidence(stdDev float64, months int) float64 {
	// Base confidence starts at 95%
	baseConfidence := 95.0

	// Reduce confidence based on historical volatility
	volatilityFactor := math.Min(stdDev/1000, 20) // Cap at 20%

	// Reduce confidence based on time horizon
	timeHorizonFactor := float64(months) * 2

	// Calculate final confidence
	confidence := baseConfidence - volatilityFactor - timeHorizonFactor

	// Ensure confidence stays between 0 and 100
	return math.Max(0, math.Min(100, confidence))
}

// generateIncomeReport genera un reporte de ingresos
func (s *CashFlowService) generateIncomeReport(movements []*models.CashFlow, period input.Period) (*input.FinancialReport, error) {
	report := &input.FinancialReport{
		Type:       input.ReportTypeIncome,
		Period:     period,
		Data:       make(map[string]float64),
		Categories: []input.CategorySummary{},
	}

	categorySums := make(map[string]float64)
	var totalIncome float64

	// Solo procesar ingresos
	for _, mov := range movements {
		if mov.OperationType.IsIncome() {
			category := string(mov.OperationType)
			categorySums[category] += mov.Amount
			totalIncome += mov.Amount
		}
	}

	// Calcular porcentajes y crear resúmenes por categoría
	for category, amount := range categorySums {
		percentage := 0.0
		if totalIncome != 0 {
			percentage = (amount / totalIncome) * 100
		}

		report.Categories = append(report.Categories, input.CategorySummary{
			Category:   category,
			Amount:     amount,
			Percentage: percentage,
		})
	}

	report.Totals = input.TotalsSummary{
		Income:    totalIncome,
		Expenses:  0,
		NetResult: totalIncome,
	}

	return report, nil
}

// generateCashFlowReport genera un reporte de flujo de caja
func (s *CashFlowService) generateCashFlowReport(movements []*models.CashFlow, period input.Period) (*input.FinancialReport, error) {
	report := &input.FinancialReport{
		Type:       input.ReportTypeCashFlow,
		Period:     period,
		Data:       make(map[string]float64),
		Categories: []input.CategorySummary{},
	}

	categorySums := make(map[string]float64)
	var totalIncome, totalExpenses float64

	// Procesar todos los movimientos
	for _, mov := range movements {
		category := string(mov.OperationType)
		categorySums[category] += mov.Amount

		if mov.OperationType.IsIncome() {
			totalIncome += mov.Amount
		} else {
			totalExpenses += mov.Amount
		}
	}

	// Calcular porcentajes y crear resúmenes por categoría
	total := math.Abs(totalIncome) + math.Abs(totalExpenses)
	for category, amount := range categorySums {
		percentage := 0.0
		if total != 0 {
			percentage = (math.Abs(amount) / total) * 100
		}

		report.Categories = append(report.Categories, input.CategorySummary{
			Category:   category,
			Amount:     amount,
			Percentage: percentage,
		})
	}

	report.Totals = input.TotalsSummary{
		Income:    totalIncome,
		Expenses:  totalExpenses,
		NetResult: totalIncome + totalExpenses,
	}

	return report, nil
}

// GetFinancialAlerts detecta y retorna alertas financieras
func (s *CashFlowService) GetFinancialAlerts(ctx context.Context) ([]input.FinancialAlert, error) {
	var alerts []input.FinancialAlert

	// Obtener balance actual
	balanceData, err := s.repository.GetBalance(ctx, nil)
	if err != nil {
		return nil, errors.DB(err, "error al obtener balance")
	}
	currentBalance := balanceData.CurrentBalance

	// Obtener movimientos del último mes
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0)

	filter := output.CashFlowFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	movements, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.DB(err, "error al obtener movimientos")
	}

	// Calcular métricas para alertas
	var (
		monthlyIncome    float64
		monthlyExpenses  float64
		unusualMovements int
	)

	// Calcular promedios y detectar movimientos inusuales
	for _, mov := range movements {
		if mov.OperationType.IsIncome() {
			monthlyIncome += mov.Amount
		} else {
			monthlyExpenses += mov.Amount
		}

		// Detectar movimientos inusuales (ejemplo: más de 5000)
		if math.Abs(mov.Amount) > 5000 {
			unusualMovements++
		}
	}

	// Alerta de balance bajo
	if currentBalance < 1000 {
		alerts = append(alerts, input.FinancialAlert{
			Type:         "balance_bajo",
			Severity:     "alta",
			Message:      "Balance actual (" + formatAmount(currentBalance) + ") por debajo del mínimo recomendado",
			Threshold:    1000,
			CurrentValue: currentBalance,
			CreatedAt:    time.Now(),
		})
	}

	// Alerta de gastos elevados
	if math.Abs(monthlyExpenses) > monthlyIncome {
		alerts = append(alerts, input.FinancialAlert{
			Type:         "gastos_elevados",
			Severity:     "media",
			Message:      "Los gastos del mes superan los ingresos",
			Threshold:    monthlyIncome,
			CurrentValue: math.Abs(monthlyExpenses),
			CreatedAt:    time.Now(),
		})
	}

	// Alerta de movimientos inusuales
	if unusualMovements > 0 {
		alerts = append(alerts, input.FinancialAlert{
			Type:         "movimientos_inusuales",
			Severity:     "baja",
			Message:      "Se detectaron " + strconv.Itoa(unusualMovements) + " movimientos de gran volumen",
			Threshold:    0,
			CurrentValue: float64(unusualMovements),
			CreatedAt:    time.Now(),
		})
	}

	// Alerta de tendencia negativa (si hay pérdidas 3 meses seguidos)
	if hasNegativeTrend := s.checkNegativeTrend(ctx); hasNegativeTrend {
		alerts = append(alerts, input.FinancialAlert{
			Type:         "tendencia_negativa",
			Severity:     "alta",
			Message:      "Tendencia negativa detectada en los últimos 3 meses",
			Threshold:    0,
			CurrentValue: currentBalance,
			CreatedAt:    time.Now(),
		})
	}

	return alerts, nil
}

// checkNegativeTrend verifica si hay una tendencia negativa en los últimos 3 meses
func (s *CashFlowService) checkNegativeTrend(ctx context.Context) bool {
	endDate := time.Now()
	startDate := endDate.AddDate(0, -3, 0)

	period := input.Period{
		StartDate: startDate,
		EndDate:   endDate,
	}

	trends, err := s.GetCashFlowTrends(ctx, period)
	if err != nil || len(trends.MonthlyTrends) < 3 {
		return false
	}

	// Verificar si los últimos 3 meses tienen crecimiento negativo
	negativeMonths := 0
	for _, trend := range trends.MonthlyTrends {
		if trend.Growth < 0 {
			negativeMonths++
		}
	}

	return negativeMonths >= 3
}
