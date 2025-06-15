package resolvers_test

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

var _ = ginkgo.Describe("CashFlow", func() {
	var (
		resolver        *resolvers.Resolver
		memberService   *test.MockMemberService
		familyService   *test.MockFamilyService
		paymentService  *test.MockPaymentService
		cashFlowService *test.MockCashFlowService
		authService     *test.MockAuthService
		userService     *test.MockUserService
	)

	ginkgo.BeforeEach(func() {
		memberService = new(test.MockMemberService)
		familyService = new(test.MockFamilyService)
		paymentService = new(test.MockPaymentService)
		cashFlowService = new(test.MockCashFlowService)
		authService = new(test.MockAuthService)
		userService = new(test.MockUserService)

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
			loginRateLimiter,
		)
	})

	ginkgo.Describe("AdjustBalance", func() {
		ginkgo.When("amount is valid", func() {
			ginkgo.It("succeeds with correct response", func() {
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

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(response).NotTo(gomega.BeNil())
				gomega.Expect(response.Success).To(gomega.BeTrue())
				gomega.Expect(response.Message).NotTo(gomega.BeNil())
				gomega.Expect(*response.Message).To(gomega.Equal("Balance adjusted successfully"))
				gomega.Expect(response.Error).To(gomega.BeNil())
			})
		})

		ginkgo.When("amount is zero", func() {
			ginkgo.It("returns validation error", func() {
				// Crear un usuario administrador para el contexto
				mockUser := &models.User{
					Role: models.RoleAdmin,
				}

				// Crear un contexto con el usuario
				ctx := context.WithValue(context.Background(), constants.UserContextKey, mockUser)

				response, err := resolver.Mutation().AdjustBalance(ctx, 0.0, "Invalid adjustment")

				// Verificaciones con la biblioteca de errores personalizada
				gomega.Expect(err).To(gomega.HaveOccurred())

				// Verificar que es un AppError
				gomega.Expect(errors.IsAppError(err)).To(gomega.BeTrue())

				// Verificar el tipo de error
				gomega.Expect(errors.Is(err, errors.ErrInvalidAmount)).To(gomega.BeTrue())

				// Verificar mensaje del error
				appErr, ok := errors.AsAppError(err)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(appErr.Message).To(gomega.ContainSubstring("cannot be zero"))

				gomega.Expect(response).To(gomega.BeNil())
			})
		})
	})
})
