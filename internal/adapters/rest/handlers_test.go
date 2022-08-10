package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTotals(t *testing.T) {
	var s Server

	t.Run(fmt.Sprintf("Totals: %d", http.StatusOK), func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/totals", nil)
		w := httptest.NewRecorder()
		s.totals(w, req)
		r := w.Result()

		if r.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d, but was %d", http.StatusOK, r.StatusCode)
		}
	})
}

func TestDelays(t *testing.T) {
	var s Server

	t.Run(fmt.Sprintf("delays: %d", http.StatusOK), func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/delays", nil)
		w := httptest.NewRecorder()
		s.delays(w, req)
		r := w.Result()
		if r.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d, but was %d", http.StatusOK, r.StatusCode)
		}
	})
}
