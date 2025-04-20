package gql_test

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors" // Importar la biblioteca de errores personalizada
	"github.com/javicabdev/asam-backend/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("CashFlow", func() {
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

	Describe("AdjustBalance", func() {
		When("amount is valid", func() {
			It("succeeds with correct response", func() {
				amount := 100.0
				reason := "Test adjustment"

				// Crear un usuario administrador para el contexto
				mockUser := &models.User{
					Role: models.RoleAdmin,
				}

				// Crear un contexto con el usuario
				ctx := context.WithValue(context.Background(), constants.UserContextKey, mockUser)

				cashFlowService.On("RegisterMovement", mock.Anything, mock.MatchedBy(func(m *models.CashFlow) bool {
					return m.Amount == amount && m.Detail == reason
				})).Return(nil)

				response, err := resolver.Mutation().AdjustBalance(ctx, amount, reason)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Success).To(BeTrue())
				Expect(response.Message).NotTo(BeNil())
				Expect(*response.Message).To(Equal("Balance adjusted successfully"))
				Expect(response.Error).To(BeNil())
			})
		})

		When("amount is zero", func() {
			It("returns validation error", func() {
				// Crear un usuario administrador para el contexto
				mockUser := &models.User{
					Role: models.RoleAdmin,
				}

				// Crear un contexto con el usuario
				ctx := context.WithValue(context.Background(), constants.UserContextKey, mockUser)

				response, err := resolver.Mutation().AdjustBalance(ctx, 0.0, "Invalid adjustment")

				// Verificaciones con la biblioteca de errores personalizada
				Expect(err).To(HaveOccurred())

				// Verificar que es un AppError
				Expect(errors.IsAppError(err)).To(BeTrue())

				// Verificar el tipo de error
				Expect(errors.Is(err, errors.ErrInvalidAmount)).To(BeTrue())

				// Verificar mensaje del error
				appErr, ok := errors.AsAppError(err)
				Expect(ok).To(BeTrue())
				Expect(appErr.Message).To(ContainSubstring("cannot be zero"))

				Expect(response).To(BeNil())
			})
		})
	})
})
