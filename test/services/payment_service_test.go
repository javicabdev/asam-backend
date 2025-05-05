package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Estos mocks no existen en el paquete test, así que debemos mantenerlos
type mockNotificationService struct {
	mock.Mock
}

func (m *mockNotificationService) SendEmail(ctx context.Context, to, subject, body string) error {
	args := m.Called(ctx, to, subject, body)
	return args.Error(0)
}

func (m *mockNotificationService) SendSMS(ctx context.Context, to, message string) error {
	args := m.Called(ctx, to, message)
	return args.Error(0)
}

type mockFeeCalculator struct {
	mock.Mock
}

func (m *mockFeeCalculator) CalculateBaseFee(year, month int) float64 {
	args := m.Called(year, month)
	return args.Get(0).(float64)
}

func (m *mockFeeCalculator) CalculateFamilyFee(year, month int) float64 {
	args := m.Called(year, month)
	return args.Get(0).(float64)
}

func (m *mockFeeCalculator) CalculateLateFee(daysLate int) float64 {
	args := m.Called(daysLate)
	return args.Get(0).(float64)
}

// Helper function to set up the test service and mocks
func setupTestService() (input.PaymentService, *test.MockPaymentRepository, *test.MockMembershipFeeRepository, *test.MockMemberRepository, *mockNotificationService, *mockFeeCalculator) {
	pr := new(test.MockPaymentRepository)
	mfr := new(test.MockMembershipFeeRepository)
	mr := new(test.MockMemberRepository)
	ns := new(mockNotificationService)
	fc := new(mockFeeCalculator)

	service := services.NewPaymentService(pr, mfr, mr, ns, fc)

	return service, pr, mfr, mr, ns, fc
}

// Helper function to create a valid payment
func createValidPayment() *models.Payment {
	return &models.Payment{
		MemberID:      1,
		Amount:        100.0,
		PaymentDate:   time.Now(),
		PaymentMethod: "efectivo",
		Status:        models.PaymentStatusPending,
	}
}

// Helper function to create a valid membership fee
func createValidMembershipFee() *models.MembershipFee {
	return &models.MembershipFee{
		Year:          time.Now().Year(),
		Month:         int(time.Now().Month()),
		BaseFeeAmount: 50.0,
		Status:        models.PaymentStatusPending,
		DueDate:       time.Now().Add(15 * 24 * time.Hour),
	}
}

// Pruebas para RegisterPayment
func TestRegisterPayment(t *testing.T) {
	// No usamos t.Parallel() en esta prueba para evitar interferencias

	tests := []struct {
		name      string
		payment   *models.Payment
		setupMock func(
			pr *test.MockPaymentRepository,
			mfr *test.MockMembershipFeeRepository,
			mr *test.MockMemberRepository,
			ns *mockNotificationService,
			fc *mockFeeCalculator,
		)
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name:    "successful registration without membership fee",
			payment: createValidPayment(),
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				// Validate member exists and is active
				// Validate member exists and is active
				member := test.CreateValidMember()
				email := "test@example.com"
				member.Email = &email
				mr.On("GetByID", mock.Anything, uint(1)).Return(member, nil)

				// Check specific parameters instead of AnythingOfType
				pr.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Payment) bool {
					return p.MemberID == 1 && p.Amount == 100.0 && p.Status == models.PaymentStatusPending
				})).Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "successful registration with membership fee",
			payment: func() *models.Payment {
				p := createValidPayment()
				feeID := uint(1)
				p.MembershipFeeID = &feeID
				return p
			}(),
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				// Validate member exists and is active
				// Validate member exists and is active
				member := test.CreateValidMember()
				email := "test@example.com"
				member.Email = &email
				mr.On("GetByID", mock.Anything, uint(1)).Return(member, nil)

				// Find membership fee
				fee := createValidMembershipFee()
				mfr.On("FindByYearMonth", mock.Anything, time.Now().Year(), int(time.Now().Month())).
					Return(fee, nil)

				// Update membership fee status
				mfr.On("Update", mock.Anything, mock.MatchedBy(func(f *models.MembershipFee) bool {
					return f.Status == models.PaymentStatusPaid
				})).Return(nil)

				// Create payment
				pr.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Payment) bool {
					return p.MemberID == 1 && p.Amount == 100.0 && p.Status == models.PaymentStatusPending
				})).Return(nil)

				// Ya no necesitamos la expectativa SendEmail, ya que el servicio no llama a este método en RegisterPayment
				// Ahora se envía en SendPaymentConfirmation que debe llamarse explícitamente
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "validation error",
			payment: &models.Payment{
				MemberID: 1,
				Amount:   -10, // Monto inválido
			},
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				// No configuration needed - will fail validation
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, appErrors.IsValidationError(err), "should be a validation error")
			},
		},
		{
			name: "membership fee not found",
			payment: func() *models.Payment {
				p := createValidPayment()
				feeID := uint(1)
				p.MembershipFeeID = &feeID
				return p
			}(),
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				// Validate member exists and is active
				mr.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)

				// Membership fee not found
				mfr.On("FindByYearMonth", mock.Anything, time.Now().Year(), int(time.Now().Month())).
					Return(nil, nil) // Fee not found
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, appErrors.IsNotFoundError(err), "should be a not found error")
			},
		},
		{
			name:    "member not found",
			payment: createValidPayment(),
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				// Member not found
				mr.On("GetByID", mock.Anything, uint(1)).Return(nil, nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, appErrors.IsNotFoundError(err), "should be a not found error")
			},
		},
		{
			name:    "inactive member",
			payment: createValidPayment(),
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				// Return inactive member
				member := test.CreateValidMember()
				member.State = models.EstadoInactivo
				mr.On("GetByID", mock.Anything, uint(1)).Return(member, nil)
				// NO AÑADIR expectativas para Create, ya que no debe llegar a ese punto
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, appErrors.IsValidationError(err), "should be a validation error")
				assert.Contains(t, err.Error(), "no está activo", "should mention inactive member")
			},
		},
		{
			name:    "database error during payment creation",
			payment: createValidPayment(),
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				member := test.CreateValidMember()
				mr.On("GetByID", mock.Anything, uint(1)).Return(member, nil)

				// Database error when creating payment
				pr.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("database error"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, appErrors.IsDatabaseError(err) || strings.Contains(err.Error(), "database error"),
					"should be a database error")
			},
		},
		{
			name: "error updating membership fee",
			payment: func() *models.Payment {
				p := createValidPayment()
				feeID := uint(1)
				p.MembershipFeeID = &feeID
				return p
			}(),
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				member := test.CreateValidMember()
				mr.On("GetByID", mock.Anything, uint(1)).Return(member, nil)

				// Find membership fee
				fee := createValidMembershipFee()
				mfr.On("FindByYearMonth", mock.Anything, time.Now().Year(), int(time.Now().Month())).
					Return(fee, nil)

				// Error updating membership fee
				mfr.On("Update", mock.Anything, mock.Anything).
					Return(errors.New("update error"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "update error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configurar un nuevo contexto con cancelación para cada test
			ctx, cancel := context.WithCancel(context.Background())
			// Asegurarse de que se cancele al finalizar la prueba
			defer cancel()

			service, pr, mfr, mr, ns, fc := setupTestService()

			tt.setupMock(pr, mfr, mr, ns, fc)

			// Ya no añadimos expectativas de mock adicionales aquí, ya que
			// las expectativas específicas se configuran en setupMock

			err := service.RegisterPayment(ctx, tt.payment)

			tt.checkErr(t, err)

			// Verificación de expectativas para cada mock
			// Verificamos solo los mocks que se esperan usar en cada test
			if tt.name != "inactive member" {
				pr.AssertExpectations(t)
			}
			mfr.AssertExpectations(t)
			mr.AssertExpectations(t)
			ns.AssertExpectations(t)
			fc.AssertExpectations(t)
		})
	}
}

// Test RegisterPayment with context cancellation
func TestRegisterPaymentWithCanceledContext(t *testing.T) {
	service, paymentRepo, _, mr, _, _ := setupTestService()

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Set up member check since this gets called first
	mr.On("GetByID", ctx, uint(1)).Return(nil, context.Canceled)

	// También configuramos Create aunque no debería llegar a llamarse
	paymentRepo.On("Create", ctx, mock.Anything).Return(context.Canceled)

	// Attempt to register a payment with canceled context
	err := service.RegisterPayment(ctx, createValidPayment())

	// The error should indicate context cancellation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// Pruebas para CancelPayment
func TestCancelPayment(t *testing.T) {
	// No usamos t.Parallel() para evitar problemas de interferencia entre tests

	tests := []struct {
		name      string
		paymentID uint
		reason    string
		setupMock func(
			pr *test.MockPaymentRepository,
			mfr *test.MockMembershipFeeRepository,
			mr *test.MockMemberRepository,
			ns *mockNotificationService,
			fc *mockFeeCalculator,
		)
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name:      "successful cancellation",
			paymentID: 1,
			reason:    "cancelado por solicitud",
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				// Configurar un payment con ID válido
				payment := &models.Payment{
					Model: gorm.Model{
						ID: 1,
					},
					MemberID:    1,
					Status:      models.PaymentStatusPending,
					PaymentDate: time.Now(),
				}
				pr.On("FindByID", mock.Anything, uint(1)).Return(payment, nil)

				// Member found for notification
				member := test.CreateValidMember()
				email := "test@example.com"
				member.Email = &email
				// No necesitamos mr.On("GetByID") porque el CancelPayment no lo usa

				// Update with correct status - usando el mismo objeto payment para garantizar integridad
				pr.On("Update", mock.Anything, mock.MatchedBy(func(p *models.Payment) bool {
					return p.ID == 1 && p.Status == models.PaymentStatusCancelled && p.Notes == "cancelado por solicitud"
				})).Return(nil)

				// No necesitamos configurar SendEmail aquí ya que no se llama durante CancelPayment
				// Solo si posteriormente se llamara a SendPaymentConfirmation
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:      "payment not found",
			paymentID: 1,
			reason:    "cancelado por solicitud",
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				pr.On("FindByID", mock.Anything, uint(1)).Return(nil, nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, appErrors.IsNotFoundError(err), "should be a not found error")
			},
		},
		{
			name:      "already cancelled payment",
			paymentID: 1,
			reason:    "cancelado nuevamente",
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				pr.On("FindByID", mock.Anything, uint(1)).
					Return(&models.Payment{
						Model: gorm.Model{
							ID: 1,
						},
						MemberID: 1,
						Status:   models.PaymentStatusCancelled,
					}, nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "El pago ya está cancelado")
			},
		},
		{
			name:      "database error when updating",
			paymentID: 1,
			reason:    "cancelado por solicitud",
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				payment := &models.Payment{
					Model: gorm.Model{
						ID: 1,
					},
					MemberID:    1,
					Status:      models.PaymentStatusPending,
					PaymentDate: time.Now(),
				}
				pr.On("FindByID", mock.Anything, uint(1)).Return(payment, nil)

				// Database error when updating
				pr.On("Update", mock.Anything, mock.MatchedBy(func(p *models.Payment) bool {
					return p.ID == 1 && p.Status == models.PaymentStatusCancelled
				})).Return(errors.New("database error"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Evitamos usar t.Parallel() para prevenir interferencias

			service, pr, mfr, mr, ns, fc := setupTestService()

			tt.setupMock(pr, mfr, mr, ns, fc)

			err := service.CancelPayment(context.Background(), tt.paymentID, tt.reason)

			tt.checkErr(t, err)

			// Verificar solo los mocks que se utilizan en CancelPayment
			pr.AssertExpectations(t)
		})
	}
}

// Pruebas para GetPayment
func TestGetPayment(t *testing.T) {
	// No usar paralelización para evitar efectos secundarios

	tests := []struct {
		name      string
		paymentID uint
		setupMock func(
			pr *test.MockPaymentRepository,
			mfr *test.MockMembershipFeeRepository,
			mr *test.MockMemberRepository,
			ns *mockNotificationService,
			fc *mockFeeCalculator,
		)
		want     *models.Payment
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name:      "successful retrieval",
			paymentID: 1,
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				expectedPayment := &models.Payment{
					Model: gorm.Model{
						ID: 1,
					},
					MemberID:    1,
					Amount:      100.0,
					Status:      models.PaymentStatusPaid,
					PaymentDate: time.Now(),
				}
				pr.On("FindByID", mock.Anything, uint(1)).Return(expectedPayment, nil)
			},
			want: &models.Payment{
				Model: gorm.Model{
					ID: 1,
				},
				MemberID:    1,
				Amount:      100.0,
				Status:      models.PaymentStatusPaid,
				PaymentDate: time.Now(),
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:      "payment not found",
			paymentID: 1,
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				pr.On("FindByID", mock.Anything, uint(1)).Return(nil, nil)
			},
			want:    nil,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, appErrors.IsNotFoundError(err), "should be a not found error")
			},
		},
		{
			name:      "database error",
			paymentID: 1,
			setupMock: func(
				pr *test.MockPaymentRepository,
				mfr *test.MockMembershipFeeRepository,
				mr *test.MockMemberRepository,
				ns *mockNotificationService,
				fc *mockFeeCalculator,
			) {
				pr.On("FindByID", mock.Anything, uint(1)).Return(nil, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Evitamos usar t.Parallel() aquí

			service, pr, _, _, _, _ := setupTestService()

			tt.setupMock(pr, nil, nil, nil, nil)

			payment, err := service.GetPayment(context.Background(), tt.paymentID)

			tt.checkErr(t, err)

			if tt.want != nil {
				assert.NotNil(t, payment)
				assert.Equal(t, tt.want.ID, payment.ID)
				assert.Equal(t, tt.want.MemberID, payment.MemberID)
				assert.Equal(t, tt.want.Amount, payment.Amount)
				assert.Equal(t, tt.want.Status, payment.Status)
			} else {
				assert.Nil(t, payment)
			}

			pr.AssertExpectations(t)
		})
	}
}

// Pruebas para GetMemberPayments
func TestGetMemberPayments(t *testing.T) {
	// No usamos t.Parallel() para evitar interferencias

	tests := []struct {
		name      string
		memberID  uint
		setupMock func(
			pr *test.MockPaymentRepository,
			mr *test.MockMemberRepository,
		)
		wantLen  int
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name:     "successful retrieval with payments",
			memberID: 1,
			setupMock: func(
				pr *test.MockPaymentRepository,
				mr *test.MockMemberRepository,
			) {
				// Member exists
				mr.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)

				// Return some payments
				payments := []models.Payment{
					{MemberID: 1, Amount: 100.0},
					{MemberID: 2, Amount: 150.0},
				}
				pr.On("FindByMember", mock.Anything, uint(1), mock.Anything, mock.Anything).
					Return(payments, nil)
			},
			wantLen: 2,
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "successful retrieval with no payments",
			memberID: 1,
			setupMock: func(
				pr *test.MockPaymentRepository,
				mr *test.MockMemberRepository,
			) {
				// Member exists
				mr.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)

				// Return empty payment list
				pr.On("FindByMember", mock.Anything, uint(1), mock.Anything, mock.Anything).
					Return([]models.Payment{}, nil)
			},
			wantLen: 0,
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "member not found",
			memberID: 1,
			setupMock: func(
				pr *test.MockPaymentRepository,
				mr *test.MockMemberRepository,
			) {
				// Member doesn't exist
				mr.On("GetByID", mock.Anything, uint(1)).Return(nil, nil)
			},
			wantLen: 0,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, appErrors.IsNotFoundError(err), "should be a not found error")
			},
		},
		{
			name:     "database error",
			memberID: 1,
			setupMock: func(
				pr *test.MockPaymentRepository,
				mr *test.MockMemberRepository,
			) {
				// Member exists
				mr.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)

				// Database error
				pr.On("FindByMember", mock.Anything, uint(1), mock.Anything, mock.Anything).
					Return([]models.Payment{}, errors.New("database error"))
			},
			wantLen: 0,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No usamos t.Parallel()

			service, pr, _, mr, _, _ := setupTestService()

			tt.setupMock(pr, mr)

			payments, err := service.GetMemberPayments(context.Background(), tt.memberID)

			tt.checkErr(t, err)

			if !tt.wantErr {
				assert.Len(t, payments, tt.wantLen)
			}

			pr.AssertExpectations(t)
			mr.AssertExpectations(t)
		})
	}
}

func TestGetFamilyPayments(t *testing.T) {
	// No usamos t.Parallel() para evitar interferencias

	tests := []struct {
		name      string
		familyID  uint
		setupMock func(
			pr *test.MockPaymentRepository,
		)
		wantLen  int
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name:     "successful retrieval with payments",
			familyID: 1,
			setupMock: func(
				pr *test.MockPaymentRepository,
			) {
				// Return some payments
				payments := []models.Payment{
					{FamilyID: &[]uint{1}[0], Amount: 100.0},
					{FamilyID: &[]uint{1}[0], Amount: 150.0},
				}
				pr.On("FindByFamily", mock.Anything, uint(1), mock.MatchedBy(func(t time.Time) bool {
					return t.Before(time.Now()) || t.Equal(time.Time{})
				}), mock.MatchedBy(func(t time.Time) bool {
					return t.After(time.Now().AddDate(-2, 0, 0)) || t.Equal(time.Now())
				})).Return(payments, nil)
			},
			wantLen: 2,
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "successful retrieval with no payments",
			familyID: 1,
			setupMock: func(
				pr *test.MockPaymentRepository,
			) {
				// Return empty payment list
				pr.On("FindByFamily", mock.Anything, uint(1), mock.Anything, mock.Anything).
					Return([]models.Payment{}, nil)
			},
			wantLen: 0,
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "database error",
			familyID: 1,
			setupMock: func(
				pr *test.MockPaymentRepository,
			) {
				// Database error
				pr.On("FindByFamily", mock.Anything, uint(1), mock.Anything, mock.Anything).
					Return([]models.Payment{}, errors.New("database error"))
			},
			wantLen: 0,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No usamos t.Parallel() para evitar interferencias

			service, pr, _, _, _, _ := setupTestService()

			tt.setupMock(pr)

			payments, err := service.GetFamilyPayments(context.Background(), tt.familyID)

			tt.checkErr(t, err)

			if !tt.wantErr {
				assert.Len(t, payments, tt.wantLen)
			}

			pr.AssertExpectations(t)
		})
	}
}

// Pruebas para GenerateMonthlyFees
func TestGenerateMonthlyFees(t *testing.T) {
	// No usamos t.Parallel() para evitar interferencias

	tests := []struct {
		name       string
		year       int
		month      int
		baseAmount float64
		setupMock  func(
			mfr *test.MockMembershipFeeRepository,
			fc *mockFeeCalculator,
		)
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name:       "successful fee generation",
			year:       2023,
			month:      5,
			baseAmount: 50.0,
			setupMock: func(
				mfr *test.MockMembershipFeeRepository,
				fc *mockFeeCalculator,
			) {
				// Check existing fee
				mfr.On("FindByYearMonth", mock.Anything, 2023, 5).Return(nil, nil)

				// Calculate family fee extra
				fc.On("CalculateFamilyFee", 2023, 5).Return(75.0)

				// Create new fee
				mfr.On("Create", mock.Anything, mock.MatchedBy(func(fee *models.MembershipFee) bool {
					return fee.Year == 2023 &&
						fee.Month == 5 &&
						fee.BaseFeeAmount == 50.0 &&
						fee.FamilyFeeExtra == 25.0
				})).Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:       "invalid base amount",
			year:       2023,
			month:      5,
			baseAmount: -10.0, // Negative amount
			setupMock: func(
				mfr *test.MockMembershipFeeRepository,
				fc *mockFeeCalculator,
			) {
				// No mocks needed - will fail validation
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, appErrors.IsValidationError(err), "should be a validation error")
			},
		},
		{
			name:       "fee already exists",
			year:       2023,
			month:      5,
			baseAmount: 50.0,
			setupMock: func(
				mfr *test.MockMembershipFeeRepository,
				fc *mockFeeCalculator,
			) {
				// Fee already exists
				mfr.On("FindByYearMonth", mock.Anything, 2023, 5).
					Return(&models.MembershipFee{
						Year:  2023,
						Month: 5,
					}, nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, appErrors.Is(err, appErrors.ErrDuplicateEntry), "should be a duplicate entry error")
			},
		},
		{
			name:       "database error",
			year:       2023,
			month:      5,
			baseAmount: 50.0,
			setupMock: func(
				mfr *test.MockMembershipFeeRepository,
				fc *mockFeeCalculator,
			) {
				// Check existing fee
				mfr.On("FindByYearMonth", mock.Anything, 2023, 5).Return(nil, nil)

				// Calculate family fee extra
				fc.On("CalculateFamilyFee", 2023, 5).Return(75.0)

				// Database error on create
				mfr.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No usamos t.Parallel() para evitar interferencias

			service, _, mfr, _, _, fc := setupTestService()

			tt.setupMock(mfr, fc)

			err := service.GenerateMonthlyFees(context.Background(), tt.year, tt.month, tt.baseAmount)

			tt.checkErr(t, err)

			mfr.AssertExpectations(t)
			fc.AssertExpectations(t)
		})
	}
}

// Pruebas de rendimiento (benchmarks)
func BenchmarkRegisterPayment(b *testing.B) {
	// Setup service and mocks
	pr := new(test.MockPaymentRepository)
	mfr := new(test.MockMembershipFeeRepository)
	mr := new(test.MockMemberRepository)
	ns := new(mockNotificationService)
	fc := new(mockFeeCalculator)

	service := services.NewPaymentService(pr, mfr, mr, ns, fc)

	// Configure mocks to avoid errors
	mr.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)
	pr.On("Create", mock.Anything, mock.Anything).Return(nil)

	// Setup payment
	payment := createValidPayment()

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.RegisterPayment(context.Background(), payment)
	}
}

func BenchmarkCancelPayment(b *testing.B) {
	// Setup service and mocks
	pr := new(test.MockPaymentRepository)
	mfr := new(test.MockMembershipFeeRepository)
	mr := new(test.MockMemberRepository)
	ns := new(mockNotificationService)
	fc := new(mockFeeCalculator)

	service := services.NewPaymentService(pr, mfr, mr, ns, fc)

	// Configure mocks to avoid errors
	payment := &models.Payment{
		Model: gorm.Model{
			ID: 1,
		},
		MemberID: 1,
		Status:   models.PaymentStatusPending,
	}
	pr.On("FindByID", mock.Anything, uint(1)).Return(payment, nil)

	mr.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)
	pr.On("Update", mock.Anything, mock.Anything).Return(nil)
	ns.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.CancelPayment(context.Background(), 1, "benchmark cancellation")
	}
}

// Test CancelPayment with context cancellation
func TestCancelPaymentWithCanceledContext(t *testing.T) {
	service, pr, _, _, _, _ := setupTestService()

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Set up FindByID to return context canceled error
	pr.On("FindByID", ctx, uint(1)).Return(nil, context.Canceled)

	// Attempt to cancel a payment with canceled context
	err := service.CancelPayment(ctx, 1, "test cancellation")

	// The error should indicate context cancellation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// Test GetPayment with context cancellation
func TestGetPaymentWithCanceledContext(t *testing.T) {
	service, pr, _, _, _, _ := setupTestService()

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Set up FindByID to return context canceled error
	pr.On("FindByID", ctx, uint(1)).Return(nil, context.Canceled)

	// Attempt to get a payment with canceled context
	payment, err := service.GetPayment(ctx, 1)

	// The error should indicate context cancellation
	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Contains(t, err.Error(), "context")
}

// Test GetMemberPayments with context cancellation
func TestGetMemberPaymentsWithCanceledContext(t *testing.T) {
	service, _, _, mr, _, _ := setupTestService()

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Set up GetByID to return context canceled error
	mr.On("GetByID", ctx, uint(1)).Return(nil, context.Canceled)

	// Attempt to get member payments with canceled context
	payments, err := service.GetMemberPayments(ctx, 1)

	// The error should indicate context cancellation
	assert.Error(t, err)
	assert.Nil(t, payments)
	assert.Contains(t, err.Error(), "context")
}

// Test GetFamilyPayments with context cancellation
func TestGetFamilyPaymentsWithCanceledContext(t *testing.T) {
	service, paymentRepo, _, _, _, _ := setupTestService()

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Set up FindByFamily to return context canceled error
	// Usar un slice vacío en lugar de nil para evitar errores de conversión
	paymentRepo.On("FindByFamily", ctx, uint(1), mock.Anything, mock.Anything).Return([]models.Payment{}, context.Canceled)

	// Attempt to get family payments with canceled context
	payments, err := service.GetFamilyPayments(ctx, 1)

	// The error should indicate context cancellation
	assert.Error(t, err)
	assert.Nil(t, payments)
	assert.Contains(t, err.Error(), "context")
}

// Test GenerateMonthlyFees with context cancellation
func TestGenerateMonthlyFeesWithCanceledContext(t *testing.T) {
	service, _, mfr, _, _, _ := setupTestService()

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Set up FindByYearMonth to return context canceled error
	mfr.On("FindByYearMonth", ctx, 2023, 5).Return(nil, context.Canceled)

	// Attempt to generate monthly fees with canceled context
	err := service.GenerateMonthlyFees(ctx, 2023, 5, 50.0)

	// The error should indicate context cancellation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}
