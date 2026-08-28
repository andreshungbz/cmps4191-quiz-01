package validator

import "testing"

// TestValidatorFlow verifies the functionality of the validator.
func TestValidatorFlow(t *testing.T) {
	v := New()

	// Assert initial valid state.
	if !v.Valid() {
		t.Fatal("new validator should start valid")
	}

	// Assert invalid state.
	v.Check(false, "name", "must be provided")
	if v.Valid() {
		t.Fatal("validator should be invalid after failed check")
	}

	// Assert error message.
	if got := v.Errors["name"]; got != "must be provided" {
		t.Fatalf("unexpected error message: got %q", got)
	}

	// Assert that adding an error with the same key does not overwrite the existing error message.
	v.AddError("name", "should not overwrite")
	if got := v.Errors["name"]; got != "must be provided" {
		t.Fatalf("error message was overwritten: got %q", got)
	}
}
