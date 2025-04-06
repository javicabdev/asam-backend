package gql_test

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

var _ = Describe("Permissions", func() {
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

	Describe("Member Operations", func() {
		When("user is admin", func() {
			It("can create member", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)
				input := model.CreateMemberInput{
					NumeroSocio:   "001",
					TipoMembresia: model.MembershipTypeIndividual,
					Nombre:        "Test",
					Apellidos:     "User",
					Direccion:     "Test Address",
					CodigoPostal:  "12345",
					Poblacion:     "Test City",
				}

				memberService.On("CreateMember", mock.Anything, mock.AnythingOfType("*models.Member")).Return(nil)

				member, err := resolver.Mutation().CreateMember(ctx, input)

				Expect(err).NotTo(HaveOccurred())
				Expect(member).NotTo(BeNil())
				memberService.AssertExpectations(GinkgoT())
			})
		})

		When("user is not admin", func() {
			BeforeEach(func() {
				memberService.ExpectedCalls = nil // Limpiamos las expectativas anteriores
			})

			It("cannot create member", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, regularUser)
				input := model.CreateMemberInput{
					NumeroSocio:   "001",
					TipoMembresia: model.MembershipTypeIndividual,
					Nombre:        "Test",
					Apellidos:     "User",
					Direccion:     "Test Address",
					CodigoPostal:  "12345",
					Poblacion:     "Test City",
				}

				member, err := resolver.Mutation().CreateMember(ctx, input)

				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, errors.ErrForbidden)).To(BeTrue())
				Expect(member).To(BeNil())
				memberService.AssertNotCalled(GinkgoT(), "CreateMember")
			})
		})
	})

	Describe("Balance Operations", func() {
		When("user is not authenticated", func() {
			It("cannot access protected endpoints", func() {
				ctx := context.Background() // Sin usuario en el contexto
				cashFlowService.On("GetCurrentBalance", mock.Anything).Return(nil, errors.New(errors.ErrUnauthorized, "no debería llegar aquí"))

				balance, err := resolver.Query().GetBalance(ctx)

				Expect(err).To(HaveOccurred())
				Expect(errors.IsAuthError(err)).To(BeTrue())
				Expect(balance).To(BeZero())
			})
		})

		When("user is admin", func() {
			It("can adjust balance", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)

				cashFlowService.On("RegisterMovement", mock.Anything, mock.AnythingOfType("*models.CashFlow")).Return(nil)

				response, err := resolver.Mutation().AdjustBalance(ctx, 100.0, "Ajuste manual")

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Success).To(BeTrue())
				cashFlowService.AssertExpectations(GinkgoT())
			})
		})

		When("user is not admin", func() {
			It("cannot adjust balance", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, regularUser)

				// No configuramos mock porque no debería llegar al servicio
				response, err := resolver.Mutation().AdjustBalance(ctx, 100.0, "Ajuste manual")

				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, errors.ErrForbidden)).To(BeTrue())
				Expect(response).To(BeNil())
				cashFlowService.AssertNotCalled(GinkgoT(), "RegisterMovement")
			})
		})
	})

	Describe("User Management", func() {
		When("user is admin", func() {
			It("can access management features", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)

				memberService.On("ListMembers", mock.Anything, mock.AnythingOfType("input.MemberFilters")).
					Return([]models.Member{}, nil)

				result, err := resolver.Query().ListMembers(ctx, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				memberService.AssertExpectations(GinkgoT())
			})
		})

		When("user is not admin", func() {
			It("cannot access management features", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, regularUser)

				result, err := resolver.Query().ListMembers(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, errors.ErrForbidden)).To(BeTrue())
				Expect(result).To(BeNil())
				memberService.AssertNotCalled(GinkgoT(), "ListMembers")
			})
		})
	})
})
