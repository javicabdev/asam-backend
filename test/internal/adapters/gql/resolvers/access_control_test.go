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

var _ = ginkgo.Describe("Access Control", func() {
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
		mockLogger               *test.MockLogger
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
		mockLogger = new(test.MockLogger)

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

	ginkgo.Describe("GetMember Access Control", func() {
		var (
			adminUser   *models.User
			regularUser *models.User
			testMember  *models.Member
		)

		ginkgo.BeforeEach(func() {
			adminUser = &models.User{
				Username: "admin",
				Role:     models.RoleAdmin,
			}
			adminUser.ID = 1

			memberID := uint(10)
			regularUser = &models.User{
				Username: "user1",
				Role:     models.RoleUser,
				MemberID: &memberID,
			}
			regularUser.ID = 2

			testMember = &models.Member{
				ID:               10,
				MembershipNumber: "M001",
				Name:             "Test Member",
			}
		})

		ginkgo.Context("when user is admin", func() {
			ginkgo.It("can access any member", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)
				memberService.On("GetMemberByID", ctx, uint(10)).Return(testMember, nil)

				result, err := resolver.Query().GetMember(ctx, "10")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.ID).To(gomega.Equal(uint(10)))
				memberService.AssertExpectations(ginkgo.GinkgoT())
			})
		})

		ginkgo.Context("when user is regular user", func() {
			ginkgo.It("can access own member", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, regularUser)
				memberService.On("GetMemberByID", ctx, uint(10)).Return(testMember, nil)

				result, err := resolver.Query().GetMember(ctx, "10")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.ID).To(gomega.Equal(uint(10)))
				memberService.AssertExpectations(ginkgo.GinkgoT())
			})

			ginkgo.It("cannot access other member", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, regularUser)

				result, err := resolver.Query().GetMember(ctx, "20")

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.Is(err, errors.ErrForbidden)).To(gomega.BeTrue())
				gomega.Expect(result).To(gomega.BeNil())
				memberService.AssertNotCalled(ginkgo.GinkgoT(), "GetMemberByID")
			})
		})

		ginkgo.Context("when user is not authenticated", func() {
			ginkgo.It("cannot access any member", func() {
				ctx := context.Background()

				result, err := resolver.Query().GetMember(ctx, "10")

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.Is(err, errors.ErrUnauthorized)).To(gomega.BeTrue())
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("ListMembers Access Control", func() {
		var (
			adminUser   *models.User
			regularUser *models.User
			userMember  *models.Member
		)

		ginkgo.BeforeEach(func() {
			adminUser = &models.User{
				Username: "admin",
				Role:     models.RoleAdmin,
			}
			adminUser.ID = 1

			memberID := uint(10)
			regularUser = &models.User{
				Username: "user1",
				Role:     models.RoleUser,
				MemberID: &memberID,
			}
			regularUser.ID = 2

			userMember = &models.Member{
				ID:   10,
				Name: "User's Member",
			}
		})

		ginkgo.Context("when user is admin", func() {
			ginkgo.It("sees all members", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)
				allMembers := []*models.Member{
					{ID: 1, Name: "Member 1"},
					{ID: 2, Name: "Member 2"},
					{ID: 3, Name: "Member 3"},
				}

				memberService.On("ListMembers", ctx, mock.AnythingOfType("input.MemberFilters")).
					Return(allMembers, nil)

				result, err := resolver.Query().ListMembers(ctx, nil)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.Nodes).To(gomega.HaveLen(3))
				gomega.Expect(result.PageInfo.TotalCount).To(gomega.Equal(3))
				memberService.AssertExpectations(ginkgo.GinkgoT())
			})
		})

		ginkgo.Context("when user is regular user", func() {
			ginkgo.It("sees only own member", func() {
				ctx := context.WithValue(context.Background(), constants.UserContextKey, regularUser)
				memberService.On("GetMemberByID", ctx, uint(10)).Return(userMember, nil)

				result, err := resolver.Query().ListMembers(ctx, nil)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.Nodes).To(gomega.HaveLen(1))
				gomega.Expect(result.PageInfo.TotalCount).To(gomega.Equal(1))
				memberService.AssertExpectations(ginkgo.GinkgoT())
			})
		})

		ginkgo.Context("when user is not authenticated", func() {
			ginkgo.It("sees nothing", func() {
				ctx := context.Background()

				result, err := resolver.Query().ListMembers(ctx, nil)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.Is(err, errors.ErrUnauthorized)).To(gomega.BeTrue())
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Mutations Require Admin", func() {
		var (
			regularUser *models.User
		)

		ginkgo.BeforeEach(func() {

			memberID := uint(10)
			regularUser = &models.User{
				Username: "user1",
				Role:     models.RoleUser,
				MemberID: &memberID,
			}
			regularUser.ID = 2
		})

		mutations := []struct {
			name     string
			mutation func(ctx context.Context) error
		}{
			{
				name: "CreateMember",
				mutation: func(ctx context.Context) error {
					_, err := resolver.Mutation().CreateMember(ctx, model.CreateMemberInput{
						NumeroSocio:     "M001",
						TipoMembresia:   model.MembershipTypeIndividual,
						Nombre:          "Test",
						Apellidos:       "User",
						CalleNumeroPiso: "Test St",
						CodigoPostal:    "12345",
						Poblacion:       "Test City",
					})
					return err
				},
			},
			{
				name: "UpdateMember",
				mutation: func(ctx context.Context) error {
					_, err := resolver.Mutation().UpdateMember(ctx, model.UpdateMemberInput{
						MiembroID: "1",
					})
					return err
				},
			},
			{
				name: "DeleteMember",
				mutation: func(ctx context.Context) error {
					_, err := resolver.Mutation().DeleteMember(ctx, "1")
					return err
				},
			},
			{
				name: "RegisterPayment",
				mutation: func(ctx context.Context) error {
					_, err := resolver.Mutation().RegisterPayment(ctx, model.PaymentInput{
						MemberID:      test.StringPtr("1"),
						Amount:        100.0,
						PaymentMethod: "cash",
					})
					return err
				},
			},
		}

		for _, tt := range mutations {
			mutation := tt // capture range variable

			ginkgo.Context("for "+mutation.name, func() {
				ginkgo.It("fails when user is regular user", func() {
					ctx := context.WithValue(context.Background(), constants.UserContextKey, regularUser)

					err := mutation.mutation(ctx)

					gomega.Expect(err).To(gomega.HaveOccurred())
					gomega.Expect(errors.Is(err, errors.ErrForbidden)).To(gomega.BeTrue())
				})

				ginkgo.It("fails when user is not authenticated", func() {
					ctx := context.Background()

					err := mutation.mutation(ctx)

					gomega.Expect(err).To(gomega.HaveOccurred())
					gomega.Expect(errors.Is(err, errors.ErrUnauthorized)).To(gomega.BeTrue())
				})
			})
		}
	})
})
