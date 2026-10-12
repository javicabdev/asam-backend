package integration

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// TestPreloadRelations guards the Preload patterns the repositories rely on
// (belongs-to, has-many, polymorphic has-many, nested, nil pointers and
// Select+Joins+Preload). It exists so a gorm bump is verified by CI instead
// of by a manual pin: gorm 1.31.1 was blocked over go-gorm/gorm#7686.
// Everything runs in a transaction that is rolled back, so it leaves no data.
func TestPreloadRelations(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	tx := database.Begin()
	if tx.Error != nil {
		t.Fatalf("begin: %v", tx.Error)
	}
	defer tx.Rollback()

	mustCreate := func(v any) {
		t.Helper()
		if err := tx.Create(v).Error; err != nil {
			t.Fatalf("create %T: %v", v, err)
		}
	}
	day := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newMember := func(number string, phones ...string) *models.Member {
		m := &models.Member{
			MembershipNumber: number, MembershipType: models.TipoMembresiaPIndividual, Name: "N", Surnames: "S",
			Address: "A", Postcode: "08001", City: "Barcelona", State: models.EstadoActivo, RegistrationDate: day,
		}
		for _, p := range phones {
			m.Telefonos = append(m.Telefonos, models.Telephone{NumeroTelefono: p})
		}
		mustCreate(m)
		return m
	}

	alone := newMember("A90001", "600000001", "600000002")
	origin := newMember("A90002", "600000003")
	noPhones := newMember("A90003")
	// A phone of another contactable type with a colliding ID must not leak in.
	mustCreate(&models.Telephone{NumeroTelefono: "699999999", ContactableID: alone.ID, ContactableType: "other"})

	family := &models.Family{
		NumeroSocio: "A91001", MiembroOrigenID: &origin.ID,
		EsposoNombre: "E", EsposoApellidos: "X", EsposaNombre: "F", EsposaApellidos: "Y",
		Telefonos: []models.Telephone{{NumeroTelefono: "700000001"}},
		Familiares: []models.Familiar{
			{Nombre: "Hijo", Apellidos: "X", Parentesco: "HIJO", Telefonos: []models.Telephone{{NumeroTelefono: "800000001"}}},
			{Nombre: "Hija", Apellidos: "X", Parentesco: "HIJA", Telefonos: []models.Telephone{{NumeroTelefono: "800000002"}, {NumeroTelefono: "800000003"}}},
		},
	}
	mustCreate(family)

	fee := &models.MembershipFee{Year: 2999, BaseFeeAmount: 30, DueDate: day}
	mustCreate(fee)
	paid := &models.Payment{MemberID: alone.ID, Amount: 30, Status: models.PaymentStatusPaid, PaymentMethod: "cash", MembershipFeeID: &fee.ID}
	feeless := &models.Payment{MemberID: noPhones.ID, Amount: 5, Status: models.PaymentStatusPending, PaymentMethod: "cash"}
	mustCreate(paid)
	mustCreate(feeless)
	linked := &models.CashFlow{MemberID: &alone.ID, PaymentID: &paid.ID, OperationType: models.OperationTypeMembershipFee, Amount: 30, Date: day, Detail: "fee"}
	orphan := &models.CashFlow{OperationType: models.OperationTypeAdministrative, Amount: -12, Date: day, Detail: "expense"}
	mustCreate(linked)
	mustCreate(orphan)
	user := &models.User{Username: "zp-preload", Email: "zp-preload@example.test", Password: "x", MemberID: &alone.ID}
	mustCreate(user)

	phones := func(ts []models.Telephone) []string {
		out := []string{}
		for _, p := range ts {
			out = append(out, p.NumeroTelefono)
		}
		sort.Strings(out)
		return out
	}
	expect := func(name string, got, want any) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}

	t.Run("member.Telefonos is polymorphic", func(t *testing.T) {
		var got []models.Member
		if err := tx.Preload("Telefonos").Where("id IN ?", []uint{alone.ID, origin.ID, noPhones.ID}).Order("id").Find(&got).Error; err != nil {
			t.Fatal(err)
		}
		expect("count", len(got), 3)
		expect("alone", phones(got[0].Telefonos), []string{"600000001", "600000002"})
		expect("origin", phones(got[1].Telefonos), []string{"600000003"})
		expect("noPhones", phones(got[2].Telefonos), []string{})
	})

	t.Run("family nested Familiares.Telefonos + MiembroOrigen + Telefonos", func(t *testing.T) {
		var got models.Family
		if err := tx.Preload("Familiares.Telefonos").Preload("MiembroOrigen").Preload("Telefonos").First(&got, family.ID).Error; err != nil {
			t.Fatal(err)
		}
		if got.MiembroOrigen == nil {
			t.Fatal("MiembroOrigen not preloaded")
		}
		expect("origin", got.MiembroOrigen.MembershipNumber, "A90002")
		expect("family phones", phones(got.Telefonos), []string{"700000001"})
		expect("familiares", len(got.Familiares), 2)
		byName := map[string][]string{}
		for _, f := range got.Familiares {
			byName[f.Nombre] = phones(f.Telefonos)
		}
		expect("familiar phones", byName, map[string][]string{"Hijo": {"800000001"}, "Hija": {"800000002", "800000003"}})
	})

	t.Run("payment belongs-to with nil fee, plain and Select+Joins", func(t *testing.T) {
		for _, joined := range []bool{false, true} {
			q := tx.Model(&models.Payment{}).Where("payments.id IN ?", []uint{paid.ID, feeless.ID})
			if joined {
				q = q.Select("payments.*").
					Joins("LEFT JOIN members ON members.id = payments.member_id").
					Joins("LEFT JOIN membership_fees ON membership_fees.id = payments.membership_fee_id")
			}
			var got []models.Payment
			if err := q.Preload("Member").Preload("MembershipFee").Order("payments.id").Find(&got).Error; err != nil {
				t.Fatal(err)
			}
			expect("count", len(got), 2)
			if got[0].Member == nil || got[0].MembershipFee == nil || got[1].Member == nil {
				t.Fatalf("joined=%v: missing preloaded association: %+v", joined, got)
			}
			expect("paid member", got[0].Member.MembershipNumber, "A90001")
			expect("paid fee", got[0].MembershipFee.Year, 2999)
			expect("feeless member", got[1].Member.MembershipNumber, "A90003")
			expect("feeless fee is nil", got[1].MembershipFee == nil, true)
		}
	})

	t.Run("cashflow belongs-to Member and Payment, nil when unset", func(t *testing.T) {
		var got []models.CashFlow
		if err := tx.Preload("Member").Preload("Payment").Where("id IN ?", []uint{linked.ID, orphan.ID}).Order("id").Find(&got).Error; err != nil {
			t.Fatal(err)
		}
		expect("count", len(got), 2)
		if got[0].Member == nil || got[0].Payment == nil {
			t.Fatalf("linked cash flow missing associations: %+v", got[0])
		}
		expect("linked member", got[0].Member.MembershipNumber, "A90001")
		expect("linked payment", got[0].Payment.ID, paid.ID)
		expect("orphan nil", got[1].Member == nil && got[1].Payment == nil, true)
	})

	t.Run("user belongs-to Member", func(t *testing.T) {
		var got models.User
		if err := tx.Preload("Member").First(&got, user.ID).Error; err != nil {
			t.Fatal(err)
		}
		if got.Member == nil {
			t.Fatal("Member not preloaded")
		}
		expect("member", got.Member.MembershipNumber, "A90001")
	})
}
