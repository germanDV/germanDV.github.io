package server

import (
	"fmt"
	"os"
	"testing"
)

const (
	UserEnvVar = "BASIC_AUTH_USER"
	PassEnvVar = "BASIC_AUTH_PASS"
)

func TestGrantPermissionWithoutSettingEnvVars(t *testing.T) {
	t.Parallel()
	match := grantPermission("bad_user", "bad_pass")
	if match {
		t.Error("Should not grant permission when not providing expected credentials via env vars")
	}
}

func TestGrantPermission(t *testing.T) {
	t.Parallel()
	correctUser := "Jose"
	correctPass := "Paquito1q2w3e4r"
	os.Setenv(UserEnvVar, correctUser)
	os.Setenv(PassEnvVar, correctPass)

	tests := []struct {
		user string
		pass string
		want bool
	}{
		{"Pepe", "NotTheRight1", false},
		{"Pepe", correctPass, false},
		{correctUser, "NotTheRight1", false},
		{correctUser, correctPass, true},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s:%s", tt.user, tt.pass)
		t.Run(testname, func(t *testing.T) {
			got := grantPermission(tt.user, tt.pass)
			if got != tt.want {
				t.Errorf("got %t, want %t (user: %s, pass: %s)", got, tt.want, tt.user, tt.pass)
			}
		})
	}
}
