package resolvers_test

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

func createPaymentAuthContext() context.Context {
	// Crear un usuario de prueba (admin para tener todos los permisos)
	testUser := &models.User{
		Username: "test_admin",
		Role:     models.RoleAdmin,
		IsActive: true,
	}
	testUser.ID = 1

	return context.WithValue(context.Background(), constants.UserContextKey, testUser)
}

var _ = ginkgo.Describe("Payment", func() {
	var (
		resolver                 *resolvers.Resolver
		memberService            *test.MockMemberService
		familyService            *test.MockFamilyService
		paymentService           *test.MockPaymentService
		cashFlowService          *test.MockCashFlowService
		authService              *test.MockAuthService
		userService              *test.MockUserService
		emailVerificationService *test.MockEmailVerificationService
		emailNotificationService *test.MockEmailNotificationService
	)

	ginkgo.BeforeEach(func() {
		memberService = new(test.MockMemberService)
		familyService = new(test.MockFamilyService)
		paymentService = new(test.MockPaymentService)
		cashFlowService = new(test.MockCashFlowService)
		authService = new(test.MockAuthService)
		userService = new(test.MockUserService)
		emailVerificationService = new(test.MockEmailVerificationService)
		emailNotificationService = new(test.MockEmailNotificationService)

		// Crear un mock logger para el rate limiter
		mockLogger := &test.MockLogger{}
		loginRateLimiter := auth.NewLoginRateLimiter(mockLogger)

		resolver = resolvers.NewResolver(
			memberService,
			familyService,
			paymentService,
			cashFlowService,
			authService,
			userService,
			emailVerificationService,
			emailNotificationService,
			loginRateLimiter,
			mockLogger,
		)
	})

	ginkgo.Describe("GetPayment", func() {
		ginkgo.When("payment exists", func() {
			ginkgo.It("returns the payment", func() {
				payment := &models.Payment{
					MemberID:      1,
					Amount:        100.0,
					PaymentMethod: "efectivo",
					Status:        models.PaymentStatusPaid,
				}

				paymentService.On("GetPayment", mock.Anything, uint(1)).Return(payment, nil)

				ctx := createPaymentAuthContext()
				result, err := resolver.Query().GetPayment(ctx, "1")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.MemberID).To(gomega.Equal(payment.MemberID))
				gomega.Expect(result.Amount).To(gomega.Equal(payment.Amount))
				gomega.Expect(result.Status).To(gomega.Equal(payment.Status))
			})
		})

		ginkgo.When("payment does not exist", func() {
			ginkgo.It("returns error", func() {
				// Aquí está la corrección: asegurarse de que el mock devuelve un error
				paymentService.On("GetPayment", mock.Anything, uint(999)).Return(nil, errors.NewNotFoundError("payment"))

				// Aquí está la corrección: usar Query().GetPayment en lugar de Mutation().CancelPayment
				ctx := createPaymentAuthContext()
				result, err := resolver.Query().GetPayment(ctx, "999")

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("RegisterPayment", func() {
		ginkgo.When("input is valid", func() {
			ginkgo.It("registers payment for active member", func() {
				input := createValidPaymentInput()

				memberService.On("GetMemberByID", mock.Anything, uint(1)).Return(&models.Member{
					ID:    1,
					State: models.EstadoActivo,
				}, nil)

				paymentService.On("RegisterPayment", mock.Anything, mock.AnythingOfType("*models.Payment")).Return(nil)

				ctx := createPaymentAuthContext()
				result, err := resolver.Mutation().RegisterPayment(ctx, input)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
			})
		})

		ginkgo.When("input is invalid", func() {
			ginkgo.It("returns error for zero amount", func() {
				input := model.PaymentInput{
					MemberID:      test.StringPtr("1"),
					Amount:        0,
					PaymentMethod: "efectivo",
				}

				ctx := createPaymentAuthContext()
				result, err := resolver.Mutation().RegisterPayment(ctx, input)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.When("member is inactive", func() {
			ginkgo.It("returns error", func() {
				input := createValidPaymentInput()

				memberService.On("GetMemberByID", mock.Anything, uint(1)).Return(&models.Member{
					ID:    1,
					State: models.EstadoInactivo,
				}, nil)

				ctx := createPaymentAuthContext()
				result, err := resolver.Mutation().RegisterPayment(ctx, input)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("CancelPayment", func() {
		ginkgo.When("payment exists", func() {
			ginkgo.It("cancels payment successfully", func() {
				paymentService.On("GetPayment", mock.Anything, uint(1)).Return(&models.Payment{
					Status: models.PaymentStatusPaid,
				}, nil)

				paymentService.On("CancelPayment", mock.Anything, uint(1), "Pago duplicado").Return(nil)

				ctx := createPaymentAuthContext()
				result, err := resolver.Mutation().CancelPayment(ctx, "1", "Pago duplicado")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.Success).To(gomega.BeTrue())
				gomega.Expect(*result.Message).To(gomega.Equal("Payment cancelled successfully"))
			})
		})

		ginkgo.When("payment does not exist", func() {
			ginkgo.It("returns error", func() {
				// Aseguramos que el servicio de pagos realmente devuelve un error
				// para un pago que no existe
				paymentService.On("GetPayment", mock.Anything, uint(999)).Return(nil, errors.NewNotFoundError("payment"))

				ctx := createPaymentAuthContext()
				result, err := resolver.Mutation().CancelPayment(ctx, "999", "test")

				// Verificamos que hay un error y que es del tipo correcto
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.IsNotFoundError(err)).To(gomega.BeTrue())

				// Verificamos que el resultado es nil
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("GenerateMonthlyFees", func() {
		ginkgo.When("parameters are valid", func() {
			ginkgo.It("generates fees successfully", func() {
				paymentService.On("GenerateMonthlyFees", mock.Anything, 2025, 1, 30.0).Return(nil)

				ctx := createPaymentAuthContext()
				result, err := resolver.Mutation().RegisterFee(ctx, 2025, 1, 30.0)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result.Success).To(gomega.BeTrue())
				gomega.Expect(*result.Message).To(gomega.Equal("Membership fee for 2025-01 has been generated successfully"))
			})
		})

		ginkgo.When("parameters are invalid", func() {
			ginkgo.It("returns error for invalid month", func() {
				// No necesitamos configurar el mock porque la validación ocurre antes de llamar al servicio
				ctx := createPaymentAuthContext()
				result, err := resolver.Mutation().RegisterFee(ctx, 2025, 13, 30.0)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.IsValidationError(err)).To(gomega.BeTrue())

				// Verificar que el campo específico tiene el error
				fields := errors.GetFields(err)
				gomega.Expect(fields).To(gomega.HaveKey("month"))

				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})
})

func createValidPaymentInput() model.PaymentInput {
	return model.PaymentInput{
		MemberID:      test.StringPtr("1"),
		Amount:        100.0,
		PaymentMethod: "efectivo",
		Notes:         test.StringPtr("Pago mensual"),
	}
}
