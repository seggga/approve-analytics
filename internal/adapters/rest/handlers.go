package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

// Handlers ...
func (s *Server) Handlers() http.Handler {
	h := chi.NewMux()
	h.Route("/", func(r chi.Router) {
		h.Use(s.CheckAuth)
		h.Get("/totals", s.totals)
		h.Get("/delays", s.delays)
	})

	return h
}

// @ID totals
// @tags analytics
// @Summary Get total counts
// @Description Get total amount of finished and declined tasks
// @Security Auth
// @Produce json
// @Success 200 {object} models.Totals true "finished and declined task counters"
// @Failure 500 {string} string "internal error"
// @Router /totals [get]
func (s *Server) totals(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("totals handler called")

	totals, _, err := s.an.GetAggregates(r.Context())
	if err != nil {
		s.logger.Sugar().Debugf("error getting aggregates %v", err)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Sugar().Debugf("got totals: %v", totals)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(totals)

	return
}

// @ID delays
// @tags analytics
// @Summary Get delays
// @Description Get delays on all finished and declined tasks
// @Security Auth
// @Produce json
// @Success 200 {array} models.Delay true "task id and lag"
// @Failure 500 {string} string "internal error"
// @Router /delays [get]
func (s *Server) delays(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("delays handler called")

	_, delays, err := s.an.GetAggregates(r.Context())
	if err != nil {
		s.logger.Sugar().Debugf("error getting aggregates %v", err)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Sugar().Debugf("got delays: %v", delays)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(delays)

	return
}
