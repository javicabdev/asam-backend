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
	"time"
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

var _ = Describe("Member", func() {
	var (
		resolver        *resolvers.Resolver
		memberService   *test.MockMemberService
		familyService   *test.MockFamilyService
		paymentService  *test.MockPaymentService
		cashFlowService *test.MockCashFlowService
		authService     *test.MockAuthService
		ctx             context.Context // Contexto autenticado para todas las pruebas
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

		// Crear contexto autenticado para todas las pruebas
		ctx = createAuthContext()
	})

	Describe("CreateMember", func() {
		When("creating individual member", func() {
			It("succeeds with valid input", func() {
				input := createValidMemberInput()

				memberService.On("CreateMember", mock.Anything, mock.MatchedBy(func(m *models.Member) bool {
					return m.TipoMembresia == models.TipoMembresiaPIndividual
				})).Return(nil)

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().CreateMember(ctx, input)

				Expect(err).NotTo(HaveOccurred())
				Expect(member).NotTo(BeNil())
				Expect(member.TipoMembresia).To(Equal(models.TipoMembresiaPIndividual))
			})
		})

		When("creating family member", func() {
			It("succeeds with valid input", func() {
				input := createValidMemberInput()
				input.TipoMembresia = model.MembershipTypeFamily

				memberService.On("CreateMember", mock.Anything, mock.MatchedBy(func(m *models.Member) bool {
					return m.TipoMembresia == models.TipoMembresiaPFamiliar
				})).Return(nil)

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().CreateMember(ctx, input)

				Expect(err).NotTo(HaveOccurred())
				Expect(member).NotTo(BeNil())
				Expect(member.TipoMembresia).To(Equal(models.TipoMembresiaPFamiliar))
			})
		})

		When("input is invalid", func() {
			It("returns error for invalid membership type", func() {
				input := createValidMemberInput()
				input.TipoMembresia = "INVALID_TYPE"

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().CreateMember(ctx, input)

				// Verificar que es un error de validación según la biblioteca de errores
				Expect(err).To(HaveOccurred())
				Expect(errors.IsValidationError(err)).To(BeTrue())

				// Obtener los campos del error
				fields := errors.GetFields(err)
				Expect(fields).To(HaveKey("tipoMembresia"))
				Expect(member).To(BeNil())
			})

			It("returns error for missing required fields", func() {
				input := model.CreateMemberInput{
					NumeroSocio: "",
					Nombre:      "",
				}

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().CreateMember(ctx, input)

				// Verificar que es un error de validación según la biblioteca de errores
				Expect(err).To(HaveOccurred())
				Expect(errors.IsValidationError(err)).To(BeTrue())

				// Obtener los campos del error
				fields := errors.GetFields(err)
				Expect(fields).To(HaveKey("numeroSocio"))
				Expect(fields).To(HaveKey("nombre"))
				Expect(member).To(BeNil())
			})
		})
	})

	Describe("UpdateMember", func() {
		When("member exists", func() {
			It("updates successfully", func() {
				existingMember := &models.Member{
					ID:            1,
					NumeroSocio:   test.GenerateValidNumeroSocio(1),
					TipoMembresia: models.TipoMembresiaPIndividual,
					Nombre:        "Juan",
					Apellidos:     "García",
					Direccion:     "Calle Test 1",
					CodigoPostal:  "08001",
					Poblacion:     "Barcelona",
					Provincia:     "Barcelona",
					Pais:          "España",
					Estado:        models.EstadoActivo,
					FechaAlta:     time.Now(),
				}

				input := model.UpdateMemberInput{
					MiembroID: "1",
					Poblacion: test.StringPtr("Nueva Ciudad"),
					Profesion: test.StringPtr("Ingeniero"),
				}

				memberService.On("GetMemberByID", mock.Anything, uint(1)).Return(existingMember, nil)
				memberService.On("UpdateMember", mock.Anything, mock.MatchedBy(func(m *models.Member) bool {
					return m.ID == 1 && m.Poblacion == "Nueva Ciudad"
				})).Return(nil)

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().UpdateMember(ctx, input)

				Expect(err).NotTo(HaveOccurred())
				Expect(member).NotTo(BeNil())
				Expect(member.Poblacion).To(Equal("Nueva Ciudad"))
				Expect(member.TipoMembresia).To(Equal(existingMember.TipoMembresia))
			})
		})

		When("member does not exist", func() {
			It("returns error", func() {
				input := model.UpdateMemberInput{
					MiembroID: "999",
					Poblacion: test.StringPtr("Ciudad Inexistente"),
				}

				memberService.On("GetMemberByID", mock.Anything, uint(999)).Return(nil, nil)

				// Usar ctx autenticado en lugar de context.Background()
				member, err := resolver.Mutation().UpdateMember(ctx, input)

				// Verificar que es un error de recurso no encontrado según la biblioteca de errores
				Expect(err).To(HaveOccurred())
				Expect(errors.IsNotFoundError(err)).To(BeTrue())

				// Verificar el mensaje del error
				message, ok := errors.GetMessage(err)
				Expect(ok).To(BeTrue())
				Expect(message).To(ContainSubstring("not found"))

				Expect(member).To(BeNil())
			})
		})
	})
})

func createValidMemberInput() model.CreateMemberInput {
	return model.CreateMemberInput{
		NumeroSocio:   "001",
		TipoMembresia: model.MembershipTypeIndividual,
		Nombre:        "Juan",
		Apellidos:     "García",
		Direccion:     "Calle Test 1",
		CodigoPostal:  "08001",
		Poblacion:     "Barcelona",
		Provincia:     test.StringPtr("Barcelona"),
		Pais:          test.StringPtr("España"),
	}
}
