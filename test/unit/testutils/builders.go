package testutils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/test/seed/generators"
)

// TestDataBuilder provides methods to build test data using the existing generators
type TestDataBuilder struct {
	rand *rand.Rand
}

// randomFirstName generates a random first name
func (b *TestDataBuilder) randomFirstName() string {
	gender := "male"
	if b.rand.Intn(2) == 0 {
		gender = "female"
	}
	firstName, _ := generators.GenerateRandomName(b.rand, gender)
	return firstName
}

// randomSurnames generates random surnames
func (b *TestDataBuilder) randomSurnames() string {
	_, lastName1 := generators.GenerateRandomName(b.rand, "male")
	_, lastName2 := generators.GenerateRandomName(b.rand, "male")
	return lastName1 + " " + lastName2
}

// randomAddress generates a random address
func (b *TestDataBuilder) randomAddress() string {
	address, _, _, _ := generators.GenerateRandomAddress(b.rand)
	return address
}

// randomPostcode generates a random postcode
func (b *TestDataBuilder) randomPostcode() string {
	_, postcode, _, _ := generators.GenerateRandomAddress(b.rand)
	return postcode
}

// randomCity generates a random city
func (b *TestDataBuilder) randomCity() string {
	_, _, city, _ := generators.GenerateRandomAddress(b.rand)
	return city
}

// randomEmail generates a random email
func (b *TestDataBuilder) randomEmail() string {
	firstName, lastName := generators.GenerateRandomName(b.rand, "male")
	firstName = strings.ToLower(strings.ReplaceAll(firstName, " ", ""))
	lastName = strings.ToLower(strings.ReplaceAll(lastName, " ", ""))
	domains := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com"}
	domain := domains[b.rand.Intn(len(domains))]
	return fmt.Sprintf("%s.%s@%s", firstName, lastName, domain)
}

// NewTestDataBuilder creates a new test data builder with a fixed seed for reproducibility
func NewTestDataBuilder() *TestDataBuilder {
	return &TestDataBuilder{
		rand: rand.New(rand.NewSource(1234)), //nolint:gosec // Intentionally using math/rand for deterministic test data
	}
}

// NewTestDataBuilderWithSeed creates a new test data builder with a custom seed
func NewTestDataBuilderWithSeed(seed int64) *TestDataBuilder {
	return &TestDataBuilder{
		rand: rand.New(rand.NewSource(seed)), //nolint:gosec // Intentionally using math/rand for deterministic test data
	}
}

// BuildValidMember creates a valid member with all required fields
func (b *TestDataBuilder) BuildValidMember() *models.Member {
	memberType := models.TipoMembresiaPIndividual
	prefix := "B"
	if b.rand.Intn(2) == 0 {
		memberType = models.TipoMembresiaPFamiliar
		prefix = "A"
	}

	return &models.Member{
		MembershipNumber: GenerateMemberNumber(prefix, b.rand.Intn(99999)+1),
		MembershipType:   memberType,
		Name:             b.randomFirstName(),
		Surnames:         b.randomSurnames(),
		Address:          b.randomAddress(),
		Postcode:         b.randomPostcode(),
		City:             b.randomCity(),
		Province:         "Barcelona",
		Country:          "España",
		State:            models.EstadoActivo,
		RegistrationDate: time.Now().AddDate(-b.rand.Intn(5), -b.rand.Intn(12), -b.rand.Intn(28)),
		Nationality:      "Senegal",
		Email:            StringPtr(b.randomEmail()),
		IdentityCard:     StringPtr(ValidSpanishDNI()),
	}
}

// BuildMemberWithNumber creates a member with a specific membership number
func (b *TestDataBuilder) BuildMemberWithNumber(memberNumber string) *models.Member {
	member := b.BuildValidMember()
	member.MembershipNumber = memberNumber

	// Adjust membership type based on prefix
	if len(memberNumber) > 0 && memberNumber[0] == 'A' {
		member.MembershipType = models.TipoMembresiaPFamiliar
	} else {
		member.MembershipType = models.TipoMembresiaPIndividual
	}

	return member
}

// BuildInactiveMember creates an inactive member with leaving date
func (b *TestDataBuilder) BuildInactiveMember() *models.Member {
	member := b.BuildValidMember()
	member.State = models.EstadoInactivo
	leavingDate := member.RegistrationDate.AddDate(1, 0, 0) // 1 year after registration
	member.LeavingDate = &leavingDate
	return member
}

// BuildValidFamily creates a valid family with all required fields
func (b *TestDataBuilder) BuildValidFamily() *models.Family {
	esposoNombre, esposoApellidos := generators.GenerateRandomName(b.rand, "male")
	esposaNombre, esposaApellidos := generators.GenerateRandomName(b.rand, "female")

	return &models.Family{
		NumeroSocio:              GenerateMemberNumber("A", b.rand.Intn(99999)+1),
		EsposoNombre:             esposoNombre,
		EsposoApellidos:          esposoApellidos,
		EsposaNombre:             esposaNombre,
		EsposaApellidos:          esposaApellidos,
		EsposoDocumentoIdentidad: ValidSpanishDNI(),
		EsposaDocumentoIdentidad: ValidSpanishNIE(),
		EsposoCorreoElectronico:  b.randomEmail(),
		EsposaCorreoElectronico:  b.randomEmail(),
	}
}

// BuildValidPayment creates a valid payment
func (b *TestDataBuilder) BuildValidPayment(memberID uint) *models.Payment {
	return &models.Payment{
		MemberID:      memberID,
		Amount:        30.0 + float64(b.rand.Intn(20)), // Between 30 and 50
		PaymentDate:   time.Now(),
		Status:        models.PaymentStatusPaid,
		PaymentMethod: "cash",
		Notes:         "Test payment",
	}
}

// BuildPendingPayment creates a pending payment
func (b *TestDataBuilder) BuildPendingPayment(memberID uint) *models.Payment {
	payment := b.BuildValidPayment(memberID)
	payment.Status = models.PaymentStatusPending
	return payment
}

// BuildValidCashFlow creates a valid cash flow entry
func (b *TestDataBuilder) BuildValidCashFlow() *models.CashFlow {
	// Generate member ID between 1-100 (safe for uint conversion)
	memberIDInt := b.rand.Intn(100) + 1
	memberID := uint(memberIDInt) //nolint:gosec // Value is guaranteed to be 1-100, safe to convert
	return &models.CashFlow{
		MemberID:      &memberID,
		OperationType: models.OperationTypeMembershipFee,
		Amount:        30.0,
		Date:          time.Now(),
		Detail:        "Cuota de membresía - Test",
	}
}

// BuildCashFlowFromPayment creates a cash flow entry from a payment
func (b *TestDataBuilder) BuildCashFlowFromPayment(payment *models.Payment) *models.CashFlow {
	return &models.CashFlow{
		MemberID:      &payment.MemberID,
		FamilyID:      payment.FamilyID,
		PaymentID:     &payment.ID,
		OperationType: models.OperationTypeMembershipFee,
		Amount:        payment.Amount,
		Date:          payment.PaymentDate,
		Detail:        "Cuota de membresía - " + payment.Notes,
	}
}

// BuildValidFamiliar creates a valid familiar (family member)
func (b *TestDataBuilder) BuildValidFamiliar(familiaID uint) *models.Familiar {
	return &models.Familiar{
		FamiliaID:         familiaID,
		Nombre:            b.randomFirstName(),
		Apellidos:         b.randomSurnames(),
		FechaNacimiento:   TimePtr(time.Now().AddDate(-b.rand.Intn(18)-5, -b.rand.Intn(12), -b.rand.Intn(28))),
		DNINIE:            ValidSpanishDNI(),
		CorreoElectronico: b.randomEmail(),
		Parentesco:        "Hijo",
	}
}

// BuildMembershipFee creates a valid membership fee
func (b *TestDataBuilder) BuildMembershipFee(year, month int) *models.MembershipFee {
	return &models.MembershipFee{
		Year:           year,
		Month:          month,
		BaseFeeAmount:  30.0,
		FamilyFeeExtra: 15.0,
		Status:         models.PaymentStatusPending,
		DueDate:        time.Date(year, time.Month(month), 10, 0, 0, 0, 0, time.UTC),
	}
}

// randomUsername generates a random username
func (b *TestDataBuilder) randomUsername() string {
	firstName, lastName := generators.GenerateRandomName(b.rand, "male")
	firstName = strings.ToLower(strings.ReplaceAll(firstName, " ", ""))
	lastName = strings.ToLower(strings.ReplaceAll(lastName, " ", ""))
	return fmt.Sprintf("%s.%s%d", firstName, lastName, b.rand.Intn(999))
}

// BuildUser creates a valid user
func (b *TestDataBuilder) BuildUser(role string) *models.User {
	return &models.User{
		Username:      b.randomUsername(),
		Email:         b.randomEmail(),
		Role:          models.Role(role),
		IsActive:      true,
		EmailVerified: true,
	}
}

// BuildUserWithMember creates a user associated with a member
func (b *TestDataBuilder) BuildUserWithMember(memberID uint) *models.User {
	user := b.BuildUser(string(models.RoleUser))
	user.MemberID = &memberID
	return user
}

// MemberBuilder provides a fluent interface for building members
type MemberBuilder struct {
	member *models.Member
}

// NewMemberBuilder creates a new member builder
func (b *TestDataBuilder) NewMemberBuilder() *MemberBuilder {
	return &MemberBuilder{
		member: b.BuildValidMember(),
	}
}

// WithMembershipNumber sets the membership number
func (mb *MemberBuilder) WithMembershipNumber(number string) *MemberBuilder {
	mb.member.MembershipNumber = number
	return mb
}

// WithName sets the member's name
func (mb *MemberBuilder) WithName(name string) *MemberBuilder {
	mb.member.Name = name
	return mb
}

// WithSurnames sets the member's surnames
func (mb *MemberBuilder) WithSurnames(surnames string) *MemberBuilder {
	mb.member.Surnames = surnames
	return mb
}

// WithState sets the member's state
func (mb *MemberBuilder) WithState(state string) *MemberBuilder {
	mb.member.State = state
	return mb
}

// WithEmail sets the member's email
func (mb *MemberBuilder) WithEmail(email string) *MemberBuilder {
	mb.member.Email = &email
	return mb
}

// WithIdentityCard sets the member's identity card
func (mb *MemberBuilder) WithIdentityCard(dni string) *MemberBuilder {
	mb.member.IdentityCard = &dni
	return mb
}

// AsInactive marks the member as inactive with a leaving date
func (mb *MemberBuilder) AsInactive() *MemberBuilder {
	mb.member.State = models.EstadoInactivo
	leavingDate := mb.member.RegistrationDate.AddDate(1, 0, 0)
	mb.member.LeavingDate = &leavingDate
	return mb
}

// AsFamily marks the member as a family type
func (mb *MemberBuilder) AsFamily() *MemberBuilder {
	mb.member.MembershipType = models.TipoMembresiaPFamiliar
	// Ensure membership number starts with 'A'
	if len(mb.member.MembershipNumber) > 0 && mb.member.MembershipNumber[0] != 'A' {
		mb.member.MembershipNumber = "A" + mb.member.MembershipNumber[1:]
	}
	return mb
}

// Build returns the constructed member
func (mb *MemberBuilder) Build() *models.Member {
	return mb.member
}
