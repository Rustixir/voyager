package voyager

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

type bindInput_ struct {
	ID    string `param:"id"`
	Tag   string `query:"tag"`
	Trace string `header:"X-Trace"`
	Note  string `json:"note"`
}

type bindOutput struct {
	ID    string `json:"id"`
	Tag   string `json:"tag"`
	Trace string `json:"trace"`
	Note  string `json:"note"`
}

func bindEcho(
	ctx *echo.Context,
	input bindInput_,
) (bindOutput, error) {
	return bindOutput{
		ID:    input.ID,
		Tag:   input.Tag,
		Trace: input.Trace,
		Note:  input.Note,
	}, nil
}

// TestBindSourcesPerMethod checks that path values, query parameters and
// headers bind for every method, while the body binds only for POST, PUT and
// PATCH.
func TestBindSourcesPerMethod(t *testing.T) {
	e := echo.New()

	handler := HandlerFunc[bindInput_, bindOutput](bindEcho)

	register(e, http.MethodGet, "/bind/:id", handler)
	register(e, http.MethodPost, "/bind/:id", handler)
	register(e, http.MethodPut, "/bind/:id", handler)
	register(e, http.MethodPatch, "/bind/:id", handler)
	register(e, http.MethodDelete, "/bind/:id", handler)

	tests := []struct {
		method   string
		wantNote string
	}{
		{http.MethodGet, ""},
		{http.MethodDelete, ""},
		{http.MethodPost, "hello"},
		{http.MethodPut, "hello"},
		{http.MethodPatch, "hello"},
	}

	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			req := httptest.NewRequest(
				test.method,
				"/bind/x_1?tag=blue",
				strings.NewReader(`{"note":"hello"}`),
			)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("X-Trace", "t_9")

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}

			body := rec.Body.String()

			// Always bound, whatever the method.
			for _, want := range []string{
				`"id":"x_1"`,
				`"tag":"blue"`,
				`"trace":"t_9"`,
			} {
				if !strings.Contains(body, want) {
					t.Errorf("missing %s in %s", want, body)
				}
			}

			// Bound only where the method defines a body.
			wantNote := `"note":"` + test.wantNote + `"`
			if !strings.Contains(body, wantNote) {
				t.Errorf("want %s, got %s", wantNote, body)
			}
		})
	}
}

func TestMethodAllowsBody(t *testing.T) {
	tests := map[string]bool{
		http.MethodPost:    true,
		http.MethodPut:     true,
		http.MethodPatch:   true,
		http.MethodGet:     false,
		http.MethodDelete:  false,
		http.MethodHead:    false,
		http.MethodOptions: false,
		http.MethodTrace:   false,
		echo.QUERY:         false,
	}

	for method, want := range tests {
		if got := methodAllowsBody(method); got != want {
			t.Errorf("methodAllowsBody(%s) = %v, want %v", method, got, want)
		}
	}
}
