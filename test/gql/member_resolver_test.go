package gql_test

import (
	"context"
	"time"

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

func createAuthContext() context.Context {
	// Crear un usuario de prueba (admin para tener todos los permisos)
	testUser := &models.User{
		Model:    gorm.Model{ID: 1},
		Username: "test_admin",
		Role:     models.RoleAdmin,
		IsActive: true,
	}

	return context.WithValue(context.Background(), constants.UserContextKey, testUser)
}

var _ = ginkgo.Describe("Member", func() {
	var (
		resolver        *resolvers.Resolver
		memberService   *test.MockMemberService
		familyService   *test.MockFamilyService
		paymentService  *test.MockPaymentService
		cashFlowService *test.MockCashFlowService
		authService     *test.MockAuthService
		ctx             context.Context // Contexto autenticado para todas las pruebas
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

		// Crear contexto autenticado para todas las pruebas
		ctx = createAuthContext()
	})

	ginkgo.Describe("CreateMember", func() {
		ginkgo.When("creating individual member", func() {
			ginkgo.It("succeeds with valid input", func() {
				input := createValidMemberInput()

				memberService.On("CreateMember", mock.Anything, mock.MatchedBy(func(m *models.Member) bool {
					return m.MembershipType == models.TipoMembresiaPIndividual
				})).Return(nil)

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().CreateMember(ctx, input)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(member).NotTo(gomega.BeNil())
				gomega.Expect(member.MembershipType).To(gomega.Equal(models.TipoMembresiaPIndividual))
			})
		})

		ginkgo.When("creating family member", func() {
			ginkgo.It("succeeds with valid input", func() {
				input := createValidMemberInput()
				input.TipoMembresia = model.MembershipTypeFamily

				memberService.On("CreateMember", mock.Anything, mock.MatchedBy(func(m *models.Member) bool {
					return m.MembershipType == models.TipoMembresiaPFamiliar
				})).Return(nil)

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().CreateMember(ctx, input)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(member).NotTo(gomega.BeNil())
				gomega.Expect(member.MembershipType).To(gomega.Equal(models.TipoMembresiaPFamiliar))
			})
		})

		ginkgo.When("input is invalid", func() {
			ginkgo.It("returns error for invalid membership type", func() {
				input := createValidMemberInput()
				input.TipoMembresia = "INVALID_TYPE"

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().CreateMember(ctx, input)

				// Verificar que es un error de validación según la biblioteca de errores
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.IsValidationError(err)).To(gomega.BeTrue())

				// Obtener los campos del error
				fields := errors.GetFields(err)
				gomega.Expect(fields).To(gomega.HaveKey("tipoMembresia"))
				gomega.Expect(member).To(gomega.BeNil())
			})

			ginkgo.It("returns error for missing required fields", func() {
				input := model.CreateMemberInput{
					NumeroSocio: "",
					Nombre:      "",
				}

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().CreateMember(ctx, input)

				// Verificar que es un error de validación según la biblioteca de errores
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.IsValidationError(err)).To(gomega.BeTrue())

				// Obtener los campos del error
				fields := errors.GetFields(err)
				gomega.Expect(fields).To(gomega.HaveKey("numeroSocio"))
				gomega.Expect(fields).To(gomega.HaveKey("nombre"))
				gomega.Expect(member).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("UpdateMember", func() {
		ginkgo.When("member exists", func() {
			ginkgo.It("updates successfully", func() {
				existingMember := &models.Member{
					ID:               1,
					MembershipNumber: test.GenerateValidNumeroSocio(1),
					MembershipType:   models.TipoMembresiaPIndividual,
					Name:             "Juan",
					Surnames:         "García",
					Address:          "Calle Test 1",
					Postcode:         "08001",
					City:             "Barcelona",
					Province:         "Barcelona",
					Country:          "España",
					State:            models.EstadoActivo,
					RegistrationDate: time.Now(),
				}

				input := model.UpdateMemberInput{
					MiembroID: "1",
					Poblacion: test.StringPtr("Nueva Ciudad"),
					Profesion: test.StringPtr("Ingeniero"),
				}

				memberService.On("GetMemberByID", mock.Anything, uint(1)).Return(existingMember, nil)
				memberService.On("UpdateMember", mock.Anything, mock.MatchedBy(func(m *models.Member) bool {
					return m.ID == 1 && m.City == "Nueva Ciudad"
				})).Return(nil)

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().UpdateMember(ctx, input)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(member).NotTo(gomega.BeNil())
				gomega.Expect(member.City).To(gomega.Equal("Nueva Ciudad"))
				gomega.Expect(member.MembershipType).To(gomega.Equal(existingMember.MembershipType))
			})
		})

		ginkgo.When("member does not exist", func() {
			ginkgo.It("returns error", func() {
				input := model.UpdateMemberInput{
					MiembroID: "999",
					Poblacion: test.StringPtr("Ciudad Inexistente"),
				}

				memberService.On("GetMemberByID", mock.Anything, uint(999)).Return(nil, nil)

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().UpdateMember(ctx, input)

				// Verificar que es un error de recurso no encontrado según la biblioteca de errores
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(errors.IsNotFoundError(err)).To(gomega.BeTrue())

				// Verificar el mensaje del error
				message, ok := errors.GetMessage(err)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(message).To(gomega.ContainSubstring("Member no encontrada"))

				gomega.Expect(member).To(gomega.BeNil())
			})
		})
	})
})

func createValidMemberInput() model.CreateMemberInput {
	return model.CreateMemberInput{
		NumeroSocio:     "001",
		TipoMembresia:   model.MembershipTypeIndividual,
		Nombre:          "Juan",
		Apellidos:       "García",
		CalleNumeroPiso: "Calle Test 1",
		CodigoPostal:    "08001",
		Poblacion:       "Barcelona",
		Provincia:       test.StringPtr("Barcelona"),
		Pais:            test.StringPtr("España"),
	}
}
