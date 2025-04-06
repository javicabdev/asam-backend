package gql_test

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Payment", func() {
	var (
		resolver        *resolvers.Resolver
		memberService   *test.MockMemberService
		familyService   *test.MockFamilyService
		paymentService  *test.MockPaymentService
		cashFlowService *test.MockCashFlowService
		authService     *test.MockAuthService
	)

	BeforeEach(func() {
		memberService = new(test.MockMemberService)
		familyService = new(test.MockFamilyService)
		paymentService = new(test.MockPaymentService)
		cashFlowService = new(test.MockCashFlowService)
		authService = new(test.MockAuthService)

		resolver = resolvers.NewResolver(
			memberService,
			familyService,
			paymentService,
			cashFlowService,
			authService,
		)
	})

	Describe("GetPayment", func() {
		When("payment exists", func() {
			It("returns the payment", func() {
				payment := &models.Payment{
					MemberID:      1,
					Amount:        100.0,
					PaymentMethod: "efectivo",
					Status:        models.PaymentStatusPaid,
				}

				paymentService.On("GetPayment", mock.Anything, uint(1)).Return(payment, nil)

				result, err := resolver.Query().GetPayment(context.Background(), "1")

				Expect(err).NotTo(HaveOccurred())
				Expect(result.MemberID).To(Equal(payment.MemberID))
				Expect(result.Amount).To(Equal(payment.Amount))
				Expect(result.Status).To(Equal(payment.Status))
			})
		})

		When("payment does not exist", func() {
			It("returns error", func() {
				paymentService.On("GetPayment", mock.Anything, uint(999)).Return(nil, errors.NewNotFoundError("payment"))

				result, err := resolver.Mutation().CancelPayment(context.Background(), "999", "test")

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("RegisterPayment", func() {
		When("input is valid", func() {
			It("registers payment for active member", func() {
				input := createValidPaymentInput()

				memberService.On("GetMemberByID", mock.Anything, uint(1)).Return(&models.Member{
					ID:     1,
					Estado: models.EstadoActivo,
				}, nil)

				paymentService.On("RegisterPayment", mock.Anything, mock.AnythingOfType("*models.Payment")).Return(nil)

				result, err := resolver.Mutation().RegisterPayment(context.Background(), input)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
			})
		})

		When("input is invalid", func() {
			It("returns error for zero amount", func() {
				input := model.PaymentInput{
					MemberID:      test.StringPtr("1"),
					Amount:        0,
					PaymentMethod: "efectivo",
				}

				result, err := resolver.Mutation().RegisterPayment(context.Background(), input)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})

		When("member is inactive", func() {
			It("returns error", func() {
				input := createValidPaymentInput()

				memberService.On("GetMemberByID", mock.Anything, uint(1)).Return(&models.Member{
					ID:     1,
					Estado: models.EstadoInactivo,
				}, nil)

				result, err := resolver.Mutation().RegisterPayment(context.Background(), input)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("CancelPayment", func() {
		When("payment exists", func() {
			It("cancels payment successfully", func() {
				paymentService.On("GetPayment", mock.Anything, uint(1)).Return(&models.Payment{
					Status: models.PaymentStatusPaid,
				}, nil)

				paymentService.On("CancelPayment", mock.Anything, uint(1), "Pago duplicado").Return(nil)

				result, err := resolver.Mutation().CancelPayment(context.Background(), "1", "Pago duplicado")

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Success).To(BeTrue())
				Expect(*result.Message).To(Equal("Payment cancelled successfully"))
			})
		})

		When("payment does not exist", func() {
			It("returns error", func() {
				paymentService.On("GetPayment", mock.Anything, uint(999)).Return(nil, errors.NewNotFoundError("payment"))

				result, err := resolver.Mutation().CancelPayment(context.Background(), "999", "test")

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("GenerateMonthlyFees", func() {
		When("parameters are valid", func() {
			It("generates fees successfully", func() {
				paymentService.On("GenerateMonthlyFees", mock.Anything, 2025, 1, 30.0).Return(nil)

				result, err := resolver.Mutation().RegisterFee(context.Background(), 2025, 1, 30.0)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Success).To(BeTrue())
				Expect(*result.Message).To(Equal("Membership fee for 2025-01 has been generated successfully"))
			})
		})

		When("parameters are invalid", func() {
			It("returns error for invalid month", func() {
				paymentService.On("GenerateMonthlyFees", mock.Anything, 2025, 13, 30.0).
					Return(errors.NewValidationError("Invalid month", nil))

				result, err := resolver.Mutation().RegisterFee(context.Background(), 2025, 13, 30.0)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
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
