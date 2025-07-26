package middleware_test

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

var _ = ginkgo.Describe("Authorization Member Middleware", func() {
	// Helper para crear punteros a uint
	uintPtr := func(i uint) *uint {
		return &i
	}

	ginkgo.Describe("GetMemberIDFromContext", func() {
		ginkgo.Context("when user is admin", func() {
			ginkgo.It("returns nil", func() {
				adminUser := &models.User{
					Model:    gorm.Model{ID: 1},
					Username: "admin",
					Role:     models.RoleAdmin,
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)

				memberID, err := middleware.GetMemberIDFromContext(ctx)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(memberID).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when user has member", func() {
			ginkgo.It("returns member ID", func() {
				user := &models.User{
					Model:    gorm.Model{ID: 2},
					Username: "user1",
					Role:     models.RoleUser,
					MemberID: uintPtr(10),
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

				memberID, err := middleware.GetMemberIDFromContext(ctx)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(memberID).NotTo(gomega.BeNil())
				gomega.Expect(*memberID).To(gomega.Equal(uint(10)))
			})
		})

		ginkgo.Context("when user has no member", func() {
			ginkgo.It("returns error", func() {
				user := &models.User{
					Model:    gorm.Model{ID: 2},
					Username: "user2",
					Role:     models.RoleUser,
					MemberID: nil,
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

				memberID, err := middleware.GetMemberIDFromContext(ctx)

				gomega.Expect(err).To(gomega.HaveOccurred())
				appErr, ok := err.(*errors.AppError)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(appErr.Code).To(gomega.Equal(errors.ErrInternalError))
				gomega.Expect(memberID).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when no user in context", func() {
			ginkgo.It("returns unauthorized error", func() {
				ctx := context.Background()

				memberID, err := middleware.GetMemberIDFromContext(ctx)

				gomega.Expect(err).To(gomega.HaveOccurred())
				appErr, ok := err.(*errors.AppError)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(appErr.Code).To(gomega.Equal(errors.ErrUnauthorized))
				gomega.Expect(memberID).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("CanAccessMember", func() {
		ginkgo.Context("when user is admin", func() {
			ginkgo.It("can access any member", func() {
				adminUser := &models.User{
					Model:    gorm.Model{ID: 1},
					Username: "admin",
					Role:     models.RoleAdmin,
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)

				err := middleware.CanAccessMember(ctx, 100)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.Context("when user is regular user", func() {
			ginkgo.It("can access own member", func() {
				user := &models.User{
					Model:    gorm.Model{ID: 2},
					Username: "user1",
					Role:     models.RoleUser,
					MemberID: uintPtr(10),
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

				err := middleware.CanAccessMember(ctx, 10)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("cannot access other member", func() {
				user := &models.User{
					Model:    gorm.Model{ID: 2},
					Username: "user1",
					Role:     models.RoleUser,
					MemberID: uintPtr(10),
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

				err := middleware.CanAccessMember(ctx, 20)

				gomega.Expect(err).To(gomega.HaveOccurred())
				appErr, ok := err.(*errors.AppError)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(appErr.Code).To(gomega.Equal(errors.ErrForbidden))
			})
		})

		ginkgo.Context("when no user in context", func() {
			ginkgo.It("returns unauthorized error", func() {
				ctx := context.Background()

				err := middleware.CanAccessMember(ctx, 10)

				gomega.Expect(err).To(gomega.HaveOccurred())
				appErr, ok := err.(*errors.AppError)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(appErr.Code).To(gomega.Equal(errors.ErrUnauthorized))
			})
		})
	})

	ginkgo.Describe("CanAccessFamily", func() {
		ginkgo.Context("when user is admin", func() {
			ginkgo.It("can access any family", func() {
				adminUser := &models.User{
					Model:    gorm.Model{ID: 1},
					Username: "admin",
					Role:     models.RoleAdmin,
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)

				err := middleware.CanAccessFamily(ctx, uintPtr(100))

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.Context("when origin member is nil", func() {
			ginkgo.It("allows access", func() {
				user := &models.User{
					Model:    gorm.Model{ID: 2},
					Username: "user1",
					Role:     models.RoleUser,
					MemberID: uintPtr(10),
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

				err := middleware.CanAccessFamily(ctx, nil)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.Context("when user is origin member", func() {
			ginkgo.It("can access family", func() {
				user := &models.User{
					Model:    gorm.Model{ID: 2},
					Username: "user1",
					Role:     models.RoleUser,
					MemberID: uintPtr(10),
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

				err := middleware.CanAccessFamily(ctx, uintPtr(10))

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.Context("when user is not origin member", func() {
			ginkgo.It("cannot access family", func() {
				user := &models.User{
					Model:    gorm.Model{ID: 2},
					Username: "user1",
					Role:     models.RoleUser,
					MemberID: uintPtr(10),
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

				err := middleware.CanAccessFamily(ctx, uintPtr(20))

				gomega.Expect(err).To(gomega.HaveOccurred())
				appErr, ok := err.(*errors.AppError)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(appErr.Code).To(gomega.Equal(errors.ErrForbidden))
			})
		})
	})

	ginkgo.Describe("CanAccessPayment", func() {
		ginkgo.Context("when user is admin", func() {
			ginkgo.It("can access any payment", func() {
				adminUser := &models.User{
					Model:    gorm.Model{ID: 1},
					Username: "admin",
					Role:     models.RoleAdmin,
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)

				err := middleware.CanAccessPayment(ctx, 100)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.Context("when payment belongs to user's member", func() {
			ginkgo.It("can access payment", func() {
				user := &models.User{
					Model:    gorm.Model{ID: 2},
					Username: "user1",
					Role:     models.RoleUser,
					MemberID: uintPtr(10),
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

				err := middleware.CanAccessPayment(ctx, 10)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.Context("when payment belongs to other member", func() {
			ginkgo.It("cannot access payment", func() {
				user := &models.User{
					Model:    gorm.Model{ID: 2},
					Username: "user1",
					Role:     models.RoleUser,
					MemberID: uintPtr(10),
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

				err := middleware.CanAccessPayment(ctx, 20)

				gomega.Expect(err).To(gomega.HaveOccurred())
				appErr, ok := err.(*errors.AppError)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(appErr.Code).To(gomega.Equal(errors.ErrForbidden))
			})
		})
	})

	ginkgo.Describe("IsUserMember", func() {
		ginkgo.It("returns true for user with member", func() {
			user := &models.User{
				Model:    gorm.Model{ID: 2},
				Username: "user1",
				Role:     models.RoleUser,
				MemberID: uintPtr(10),
			}
			ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

			result := middleware.IsUserMember(ctx)

			gomega.Expect(result).To(gomega.BeTrue())
		})

		ginkgo.It("returns false for user without member", func() {
			user := &models.User{
				Model:    gorm.Model{ID: 2},
				Username: "user1",
				Role:     models.RoleUser,
				MemberID: nil,
			}
			ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

			result := middleware.IsUserMember(ctx)

			gomega.Expect(result).To(gomega.BeFalse())
		})

		ginkgo.It("returns false for admin", func() {
			adminUser := &models.User{
				Model:    gorm.Model{ID: 1},
				Username: "admin",
				Role:     models.RoleAdmin,
			}
			ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)

			result := middleware.IsUserMember(ctx)

			gomega.Expect(result).To(gomega.BeFalse())
		})

		ginkgo.It("returns false when no user in context", func() {
			ctx := context.Background()

			result := middleware.IsUserMember(ctx)

			gomega.Expect(result).To(gomega.BeFalse())
		})
	})

	ginkgo.Describe("GetCurrentUserMember", func() {
		ginkgo.Context("when user has member", func() {
			ginkgo.It("returns member ID", func() {
				user := &models.User{
					Model:    gorm.Model{ID: 2},
					Username: "user1",
					Role:     models.RoleUser,
					MemberID: uintPtr(10),
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, user)

				memberID, err := middleware.GetCurrentUserMember(ctx)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(memberID).NotTo(gomega.BeNil())
				gomega.Expect(*memberID).To(gomega.Equal(uint(10)))
			})
		})

		ginkgo.Context("when user is admin", func() {
			ginkgo.It("returns nil", func() {
				adminUser := &models.User{
					Model:    gorm.Model{ID: 1},
					Username: "admin",
					Role:     models.RoleAdmin,
				}
				ctx := context.WithValue(context.Background(), constants.UserContextKey, adminUser)

				memberID, err := middleware.GetCurrentUserMember(ctx)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(memberID).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when no user in context", func() {
			ginkgo.It("returns error", func() {
				ctx := context.Background()

				memberID, err := middleware.GetCurrentUserMember(ctx)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(memberID).To(gomega.BeNil())
			})
		})
	})
})
