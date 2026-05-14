package characterization_test

import (
	"backend-v2/internal/dto"
	"backend-v2/test/characterization/helpers"
	"strings"
	"testing"
)

func TestLogin_HappyPath_Project1(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}

	env := helpers.RawJSON(t, "POST", "/user/login", dto.LoginData{
		Username:  helpers.TesterUsername,
		Password:  helpers.TesterPassword,
		ProjectID: 1,
	})
	helpers.AssertSuccess(t, env, "login project 1")

	var lr dto.LoginResponse
	helpers.MustDecode(t, env, &lr)
	if lr.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if lr.Admin {
		t.Fatal("project 1 should not set admin=true")
	}

	helpers.AssertJSONGolden(t, "auth/login_project1", env.Data)
}

func TestLogin_AdminProject_SetsAdminTrue(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}

	env := helpers.RawJSON(t, "POST", "/user/login", dto.LoginData{
		Username:  helpers.TesterUsername,
		Password:  helpers.TesterPassword,
		ProjectID: 2,
	})
	helpers.AssertSuccess(t, env, "login project 2")

	var lr dto.LoginResponse
	helpers.MustDecode(t, env, &lr)
	if !lr.Admin {
		t.Fatal("project 2 (Администрирование) should set admin=true")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}

	env := helpers.RawJSON(t, "POST", "/user/login", dto.LoginData{
		Username:  helpers.TesterUsername,
		Password:  "obviously-wrong",
		ProjectID: 1,
	})
	msg := helpers.AssertFailure(t, env, "login wrong password")
	// Outer wrap: user_controller.Login → "Ошибка при входе: <inner>"
	// Inner: user_service.Login → "Неправильный пароль"
	if !strings.Contains(msg, "Неправильный пароль") {
		t.Fatalf("expected wrong-password error, got %q", msg)
	}
}

func TestLogin_NoAccessToProject(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}
	// Default tester has the Суперадмин role, which bypasses the project
	// access check. Use a regular-role user so we actually exercise the
	// "not in user_in_projects for this project" path.
	if err := helpers.SeedRegularUser(); err != nil {
		t.Fatalf("SeedRegularUser: %v", err)
	}

	env := helpers.RawJSON(t, "POST", "/user/login", dto.LoginData{
		Username:  helpers.RegularUserUsername,
		Password:  helpers.RegularUserPassword,
		ProjectID: 999,
	})
	msg := helpers.AssertFailure(t, env, "login no access")
	if !strings.Contains(msg, "У вас нету доступа в выбранный проект") {
		t.Fatalf("expected no-access-to-project error, got %q", msg)
	}
}


func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}

	env := helpers.RawJSON(t, "GET", "/input/paginated", nil)
	msg := helpers.AssertFailure(t, env, "missing auth header")
	if msg != "Ошибка идентификации: вы не являетесь пользователем" {
		t.Fatalf("unexpected error message: %q", msg)
	}
}

func TestAuthMiddleware_NonBearerScheme(t *testing.T) {
	if err := helpers.ResetDB(); err != nil {
		t.Fatalf("ResetDB: %v", err)
	}

	env := helpers.RequestJSON(t, "GET", "/input/paginated",
		map[string]string{"Authorization": "Basic xxx"}, nil)
	msg := helpers.AssertFailure(t, env, "non-bearer scheme")
	if msg != "Ошибка идентификации: неправильная аутентификация" {
		t.Fatalf("unexpected error message: %q", msg)
	}
}
