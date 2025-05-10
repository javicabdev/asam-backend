package gql_test

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

var _ = ginkgo.Describe("Family", func() {
	var (
		resolver        *resolvers.Resolver
		memberService   *test.MockMemberService
		familyService   *test.MockFamilyService
		paymentService  *test.MockPaymentService
		cashFlowService *test.MockCashFlowService
		authService     *test.MockAuthService
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
	})

	ginkgo.Describe("GetFamily", func() {
		ginkgo.When("family exists", func() {
			ginkgo.It("returns the family", func() {
				expectedFamily := &models.Family{
					ID:              1,
					NumeroSocio:     "A0001",
					EsposoNombre:    "Juan",
					EsposoApellidos: "García",
				}

				familyService.On("GetByID", mock.Anything, uint(1)).Return(expectedFamily, nil)

				family, err := resolver.Query().GetFamily(context.Background(), "1")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(family.ID).To(gomega.Equal(expectedFamily.ID))
				gomega.Expect(family.NumeroSocio).To(gomega.Equal(expectedFamily.NumeroSocio))
				gomega.Expect(family.EsposoNombre).To(gomega.Equal(expectedFamily.EsposoNombre))
				gomega.Expect(family.EsposoApellidos).To(gomega.Equal(expectedFamily.EsposoApellidos))
			})
		})

		ginkgo.When("family does not exist", func() {
			ginkgo.It("returns not found error", func() {
				familyService.On("GetByID", mock.Anything, uint(999)).Return(nil, appErrors.NewNotFoundError("familia"))

				family, err := resolver.Query().GetFamily(context.Background(), "999")

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(appErrors.IsNotFoundError(err)).To(gomega.BeTrue(), "Debería ser un error de tipo 'no encontrado'")
				gomega.Expect(family).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("CreateFamily", func() {
		ginkgo.When("input is valid", func() {
			ginkgo.It("creates the family", func() {
				input := createValidFamilyInput()

				familyService.On("Create", mock.Anything, mock.AnythingOfType("*models.Family")).Return(nil)

				family, err := resolver.Mutation().CreateFamily(context.Background(), input)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(family).NotTo(gomega.BeNil())
			})
		})

		ginkgo.When("input is invalid", func() {
			ginkgo.It("returns validation error", func() {
				input := model.CreateFamilyInput{
					NumeroSocio:  "",
					EsposoNombre: "",
				}

				// Simular un error de validación
				expectedErr := appErrors.Validation("El número de familia es obligatorio", "numeroSocio", "El número de familia es obligatorio")

				familyService.On("Create", mock.Anything, mock.AnythingOfType("*models.Family")).Return(expectedErr)

				family, err := resolver.Mutation().CreateFamily(context.Background(), input)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(appErrors.IsValidationError(err)).To(gomega.BeTrue(), "Debería ser un error de validación")

				// Verificar que contiene errores en los campos esperados
				fields := appErrors.GetFields(err)
				gomega.Expect(fields).To(gomega.HaveKey("numeroSocio"), "Debería tener error en el campo 'numeroSocio'")
				gomega.Expect(family).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("AddFamilyMember", func() {
		ginkgo.When("family exists", func() {
			ginkgo.It("adds the family member", func() {
				familyID := "1"
				fechaNacimiento := time.Now()
				familiar := model.FamiliarInput{
					Nombre:          "Pedro",
					Apellidos:       "García López",
					FechaNacimiento: &fechaNacimiento,
					DniNie:          test.StringPtr("12345678C"),
					Parentesco:      "Hijo",
				}

				familyService.On("GetByID", mock.Anything, uint(1)).Return(&models.Family{
					ID:          1,
					NumeroSocio: "A0001",
				}, nil)
				familyService.On("AddFamiliar", mock.Anything, uint(1), mock.AnythingOfType("*models.Familiar")).Return(nil)

				member, err := resolver.Mutation().AddFamilyMember(context.Background(), familyID, familiar)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(member).NotTo(gomega.BeNil())
			})
		})

		ginkgo.When("family does not exist", func() {
			ginkgo.It("returns not found error", func() {
				familyID := "999"
				familiar := model.FamiliarInput{
					Nombre:     "Pedro",
					Apellidos:  "García",
					Parentesco: "Hijo",
				}

				familyService.On("GetByID", mock.Anything, uint(999)).Return(nil, appErrors.NewNotFoundError("familia"))

				member, err := resolver.Mutation().AddFamilyMember(context.Background(), familyID, familiar)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(appErrors.IsNotFoundError(err)).To(gomega.BeTrue(), "Debería ser un error de tipo 'no encontrado'")
				gomega.Expect(member).To(gomega.BeNil())
			})
		})
	})
})

func createValidFamilyInput() model.CreateFamilyInput {
	esposoFecha := time.Now().AddDate(-30, 0, 0)
	esposaFecha := time.Now().AddDate(-28, 0, 0)
	return model.CreateFamilyInput{
		NumeroSocio:              "A0001",
		EsposoNombre:             "Juan",
		EsposoApellidos:          "García",
		EsposaNombre:             "María",
		EsposaApellidos:          "López",
		EsposoFechaNacimiento:    &esposoFecha,
		EsposaFechaNacimiento:    &esposaFecha,
		EsposoDocumentoIdentidad: test.StringPtr("12345678A"),
		EsposoCorreoElectronico:  test.StringPtr("juan@example.com"),
		EsposaDocumentoIdentidad: test.StringPtr("87654321B"),
		EsposaCorreoElectronico:  test.StringPtr("maria@example.com"),
	}
}
