package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	port      int
	lastPhase string
	roundCh   chan<- Round
}

func NewServer(port int, roundCh chan<- Round) *Server {
	return &Server{
		port:    port,
		roundCh: roundCh,
	}
}

func (s *Server) handleGameState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	var state GameState
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		return
	}

	if state.Round.Phase != "" && state.Round.Phase != s.lastPhase {
		s.lastPhase = state.Round.Phase
		s.roundCh <- state.Round
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) StartListener() {
	http.HandleFunc("/", s.handleGameState)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil))
}
