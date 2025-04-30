package auth

import "testing"

func TestGetUser(t *testing.T) {
	expected := "DefaultUser"
	if got := GetUser(); got !=
		expected {
		t.Errorf("GetUser() = %q, want %q", got, expected)
	}
}
