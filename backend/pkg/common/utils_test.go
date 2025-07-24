package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bmizerany/assert"
)

func TestSetContextValuesMiddleware(t *testing.T) {
	key1 := ContextKey("userID")
	key2 := ContextKey("role")
	key3 := ContextKey("slicemap")
	users := []User{
		{ID: 3, Roles: []string{"r1"}},
		{ID: 4, Roles: []string{"r4", "r99"}},
	}

	middleware := SetContextValuesMiddleware(map[ContextKey]any{
		key1: 1234,
		key2: "admin",
		key3: users,
	})

	var gotValues map[ContextKey]any

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotValues = map[ContextKey]any{
			key1: r.Context().Value(key1),
			key2: r.Context().Value(key2),
			key3: r.Context().Value(key3),
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, 1234, gotValues[key1])
	assert.Equal(t, "admin", gotValues[key2])
	assert.Equal(t, users, gotValues[key3])
}
