package voyager

import (
	"context"
	"github.com/labstack/echo/v5"
	"net/http/httptest"
	"strings"
	"testing"
)

type User struct{}
type UserRepository interface {
	FindByID(ctx context.Context, ID string) (User, error)
}

type UserSetter interface {
	SetUser(User)
}

type ProvideUserInterceptor struct {
	repo UserRepository
}

func (i ProvideUserInterceptor) Intercept(
	ctx *echo.Context,
	input any,
) error {
	setter, ok := input.(UserSetter)
	if !ok {
		return nil
	}

	userID := ctx.Param("userId")

	user, err := i.repo.FindByID(
		ctx.Request().Context(),
		userID,
	)
	if err != nil {
		return err
	}

	setter.SetUser(user)

	return nil
}

func TestInterceptFunctionsBeforeValidation(t *testing.T) {
	app := New("Interceptors", "1.0.0")
	var order []int
	app.Use(Intercept(
		func(ctx *echo.Context, input any) error {
			order = append(order, 1)
			input.(*createInput).Name = "default"
			return nil
		},
		func(ctx *echo.Context, input any) error {
			order = append(order, 2)
			return nil
		},
	))
	POST(app, "/interceptor-functions/:id", createHandler)
	req := httptest.NewRequest("POST", "/interceptor-functions/42", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"name":"default"`) {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Fatalf("unexpected interceptor order: %v", order)
	}
}
