package testutils

import (
	"reflect"
	"strings"
	"testing"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// AssertValidationError asserts that the error is a validation error and optionally checks the field
func AssertValidationError(t *testing.T, err error, expectedField string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected validation error but got nil")
	}

	appErr, ok := errors.AsAppError(err)
	if !ok {
		t.Fatalf("expected AppError but got %T: %v", err, err)
	}

	if appErr.Code != errors.ErrValidationFailed {
		t.Fatalf("expected validation error code but got %s", appErr.Code)
	}

	if expectedField != "" {
		if appErr.Fields == nil || appErr.Fields[expectedField] == "" {
			t.Fatalf("expected error on field %s but field not found in error fields", expectedField)
		}
	}
}

// AssertMemberValid asserts that a member passes all validations
func AssertMemberValid(t *testing.T, member *models.Member) {
	t.Helper()

	if err := member.Validate(); err != nil {
		t.Fatalf("expected member to be valid but got error: %v", err)
	}
}

// AssertMemberInvalid asserts that a member fails validation
func AssertMemberInvalid(t *testing.T, member *models.Member) {
	t.Helper()

	if err := member.Validate(); err == nil {
		t.Fatal("expected member to be invalid but validation passed")
	}
}

// AssertMemberState asserts that a member has the expected state
func AssertMemberState(t *testing.T, member *models.Member, expectedState string) {
	t.Helper()

	if member.State != expectedState {
		t.Fatalf("expected member state %s but got %s", expectedState, member.State)
	}
}

// AssertMemberType asserts that a member has the expected membership type
func AssertMemberType(t *testing.T, member *models.Member, expectedType string) {
	t.Helper()

	if member.MembershipType != expectedType {
		t.Fatalf("expected membership type %s but got %s", expectedType, member.MembershipType)
	}
}

// AssertPaymentValid asserts that a payment passes all validations
func AssertPaymentValid(t *testing.T, payment *models.Payment) {
	t.Helper()

	if err := payment.Validate(); err != nil {
		t.Fatalf("expected payment to be valid but got error: %v", err)
	}
}

// AssertPaymentStatus asserts that a payment has the expected status
func AssertPaymentStatus(t *testing.T, payment *models.Payment, expectedStatus models.PaymentStatus) {
	t.Helper()

	if payment.Status != expectedStatus {
		t.Fatalf("expected payment status %s but got %s", expectedStatus, payment.Status)
	}
}

// AssertCashFlowValid asserts that a cash flow entry passes all validations
func AssertCashFlowValid(t *testing.T, cashFlow *models.CashFlow) {
	t.Helper()

	if err := cashFlow.Validate(); err != nil {
		t.Fatalf("expected cash flow to be valid but got error: %v", err)
	}
}

// AssertFamilyValid asserts that a family passes all validations
func AssertFamilyValid(t *testing.T, family *models.Family) {
	t.Helper()

	if err := family.Validate(); err != nil {
		t.Fatalf("expected family to be valid but got error: %v", err)
	}
}

// AssertEqualMembers asserts that two members are equal (ignoring timestamps)
func AssertEqualMembers(t *testing.T, expected, actual *models.Member) {
	t.Helper()

	// Compare all fields except timestamps
	if expected.MembershipNumber != actual.MembershipNumber {
		t.Errorf("MembershipNumber mismatch: expected %s, got %s", expected.MembershipNumber, actual.MembershipNumber)
	}
	if expected.MembershipType != actual.MembershipType {
		t.Errorf("MembershipType mismatch: expected %s, got %s", expected.MembershipType, actual.MembershipType)
	}
	if expected.Name != actual.Name {
		t.Errorf("Name mismatch: expected %s, got %s", expected.Name, actual.Name)
	}
	if expected.Surnames != actual.Surnames {
		t.Errorf("Surnames mismatch: expected %s, got %s", expected.Surnames, actual.Surnames)
	}
	if expected.State != actual.State {
		t.Errorf("State mismatch: expected %s, got %s", expected.State, actual.State)
	}
	// Add more field comparisons as needed
}

// AssertErrorContains asserts that an error contains a specific substring
func AssertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing '%s' but got nil", substr)
	}

	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("expected error to contain '%s' but got: %v", substr, err)
	}
}

// AssertErrorType asserts that an error is of a specific type
func AssertErrorType(t *testing.T, err error, expectedType interface{}) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error of type %T but got nil", expectedType)
	}

	errType := reflect.TypeOf(err)
	expectedErrType := reflect.TypeOf(expectedType)

	if errType != expectedErrType {
		t.Fatalf("expected error of type %v but got %v", expectedErrType, errType)
	}
}

// AssertNil asserts that a value is nil
func AssertNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if value != nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("expected nil but got %v - %v", value, msgAndArgs)
		} else {
			t.Fatalf("expected nil but got %v", value)
		}
	}
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if value == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("expected non-nil value - %v", msgAndArgs)
		} else {
			t.Fatal("expected non-nil value")
		}
	}
}

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("values are not equal:\nexpected: %+v\nactual:   %+v", expected, actual)
	}
}

// AssertNotEqual asserts that two values are not equal
func AssertNotEqual(t *testing.T, v1, v2 interface{}) {
	t.Helper()

	if reflect.DeepEqual(v1, v2) {
		t.Fatalf("values should not be equal but both are: %+v", v1)
	}
}

// AssertTrue asserts that a boolean value is true
func AssertTrue(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()

	if !value {
		if len(msgAndArgs) > 0 {
			t.Fatalf("expected true but got false - %v", msgAndArgs)
		} else {
			t.Fatal("expected true but got false")
		}
	}
}

// AssertFalse asserts that a boolean value is false
func AssertFalse(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()

	if value {
		if len(msgAndArgs) > 0 {
			t.Fatalf("expected false but got true - %v", msgAndArgs)
		} else {
			t.Fatal("expected false but got true")
		}
	}
}
