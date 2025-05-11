package resolvers_test

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

var _ = ginkgo.Describe("Permissions", func() {
	var (
		resolver        *resolvers.Resolver
		memberService   *test.MockMemberService
		familyService   *test.MockFamilyService
		paymentService  *test.MockPaymentService
		cashFlowService *test.MockCashFlowService
		authService     *test.MockAuthService
		adminUser       *models.User
		regularUser     *models.User
	)

	ginkgo.BeforeEach(func() {
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

		adminUser = &models.User{
			Model:    gorm.Model{ID: 1},
			Username: "admin",
			Role:     models.RoleAdmin,
		}

		regularUser = &models.User{
			Model:    gorm.Model{ID: 2},
			Username: "user",
			Role:     models.RoleUser,
		}
	})

	ginkgo.Describe("Member Operations", func() {
		ginkgo.When("user is admin", func() {
			ginkgo.It("can create member", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)
				input := model.CreateMemberInput{
					NumeroSocio:     "001",
					TipoMembresia:   model.MembershipTypeIndividual,
					Nombre:          "Test",
					Apellidos:       "User",
					CalleNumeroPiso: "Test Address",
					CodigoPostal:    "12345",
					Poblacion:       "Test City",
				}

				memberService.On("CreateMember", mock.Anything, mock.AnythingOfType("*models.Member")).Return(nil)

				member, err := resolver.Mutation().CreateMember(ctx, input)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(member).NotTo(gomega.BeNil())
				memberService.AssertExpectations(ginkgo.GinkgoT())
			})
		})

		ginkgo.When("user is not admin", func() {
			ginkgo.BeforeEach(func() {
				memberService.ExpectedCalls = nil // Limpiamos las expectativas anteriores
			})

			ginkgo.It("cannot create member", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, regularUser)
				input := model.CreateMemberInput{
					NumeroSocio:     "001",
					TipoMembresia:   model.MembershipTypeIndividual,
					Nombre:          "Test",
					Apellidos:       "User",
					CalleNumeroPiso: "Test Address",
					CodigoPostal:    "12345",
					Poblacion:       "Test City",
				}

				member, err := resolver.Mutation().CreateMember(ctx, input)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.Is(err, errors.ErrForbidden)).To(gomega.BeTrue(), "Debería ser un error de permisos insuficientes")
				gomega.Expect(member).To(gomega.BeNil())
				memberService.AssertNotCalled(ginkgo.GinkgoT(), "CreateMember")
			})
		})
	})

	ginkgo.Describe("Balance Operations", func() {
		ginkgo.When("user is not authenticated", func() {
			ginkgo.It("cannot access protected endpoints", func() {
				ctx := context.Background() // Sin usuario en el contexto
				cashFlowService.On("GetCurrentBalance", mock.Anything).Return(nil, errors.New(errors.ErrUnauthorized, "no debería llegar aquí"))

				balance, err := resolver.Query().GetBalance(ctx)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.IsAuthError(err)).To(gomega.BeTrue(), "Debería ser un error de autenticación")
				gomega.Expect(balance).To(gomega.BeZero())
			})
		})

		ginkgo.When("user is admin", func() {
			ginkgo.It("can adjust balance", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)

				cashFlowService.On("RegisterMovement", mock.Anything, mock.AnythingOfType("*models.CashFlow")).Return(nil)

				response, err := resolver.Mutation().AdjustBalance(ctx, 100.0, "Ajuste manual")

				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "No debería haber error para un usuario admin")
				gomega.Expect(response).NotTo(gomega.BeNil())
				gomega.Expect(response.Success).To(gomega.BeTrue())
				cashFlowService.AssertExpectations(ginkgo.GinkgoT())
			})
		})

		ginkgo.When("user is not admin", func() {
			ginkgo.It("cannot adjust balance", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, regularUser)

				// Configuramos el mock para capturar cualquier llamada
				cashFlowService.On("RegisterMovement", mock.Anything, mock.AnythingOfType("*models.CashFlow")).Return(errors.NewBusinessError(errors.ErrForbidden, "Insufficient permissions"))

				response, err := resolver.Mutation().AdjustBalance(ctx, 100.0, "Ajuste manual")

				gomega.Expect(err).To(gomega.HaveOccurred(), "Debería haber error para un usuario no admin")
				gomega.Expect(errors.Is(err, errors.ErrForbidden)).To(gomega.BeTrue(), "Debería ser un error de tipo 'acceso prohibido'")
				gomega.Expect(response).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("User Management", func() {
		ginkgo.When("user is admin", func() {
			ginkgo.It("can access management features", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)

				// Cambiamos el tipo de retorno a []*models.Member en lugar de []models.Member
				members := make([]*models.Member, 0)
				memberService.On("ListMembers", mock.Anything, mock.AnythingOfType("input.MemberFilters")).
					Return(members, nil)

				result, err := resolver.Query().ListMembers(ctx, nil)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				memberService.AssertExpectations(ginkgo.GinkgoT())
			})
		})

		ginkgo.When("user is not admin", func() {
			ginkgo.It("cannot access management features", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, regularUser)

				// Configuramos el mock para capturar cualquier llamada
				members := make([]*models.Member, 0)
				memberService.On("ListMembers", mock.Anything, mock.AnythingOfType("input.MemberFilters")).
					Return(members, errors.NewBusinessError(errors.ErrForbidden, "Insufficient permissions"))

				result, err := resolver.Query().ListMembers(ctx, nil)

				gomega.Expect(err).To(gomega.HaveOccurred(), "Debería haber error para un usuario no admin")
				gomega.Expect(errors.Is(err, errors.ErrForbidden)).To(gomega.BeTrue(), "Debería ser un error de tipo 'acceso prohibido'")
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})
})
