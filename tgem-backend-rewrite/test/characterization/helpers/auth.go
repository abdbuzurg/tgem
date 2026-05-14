package helpers

import (
	"backend-v2/internal/dto"
	"testing"
)

// Login posts to /user/login and returns the JWT. Asserts envelope.success=true.
func Login(t *testing.T, username, password string, projectID uint) string {
	t.Helper()
	env := RawJSON(t, "POST", "/user/login", dto.LoginData{
		Username:  username,
		Password:  password,
		ProjectID: projectID,
	})
	AssertSuccess(t, env, "Login")

	var lr dto.LoginResponse
	MustDecode(t, env, &lr)
	if lr.Token == "" {
		t.Fatalf("Login: empty token")
	}
	return lr.Token
}

// LoginAsTester logs in with the seeded tester user against project 1.
func LoginAsTester(t *testing.T) string {
	t.Helper()
	return Login(t, TesterUsername, TesterPassword, 1)
}
