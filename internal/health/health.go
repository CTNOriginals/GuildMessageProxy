package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/CTNOriginals/GuildMessageProxy/internal/commands"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
)

// HealthStatus represents the current health state of the bot.
type HealthStatus struct {
	Status         string        `json:"status"`
	BotConnected   bool          `json:"bot_connected"`
	CommandCount   int           `json:"command_count"`
	Uptime         time.Duration `json:"uptime"`
	Timestamp      time.Time     `json:"timestamp"`
}

// Server provides HTTP health check endpoints for the bot.
type Server struct {
	httpServer *http.Server
	bot        *discordgo.Session
	startTime  time.Time
	mux        *http.ServeMux
}

// NewServer creates a new health check server.
func NewServer(bot *discordgo.Session, startTime time.Time) *Server {
	var mux = http.NewServeMux()
	var s = &Server{
		bot:       bot,
		startTime: startTime,
		mux:       mux,
	}

	// Register handlers
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)
	mux.HandleFunc("/live", s.handleLive)

	return s
}

// handleHealth returns full health status as JSON.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var status = s.getHealthStatus()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var err = json.NewEncoder(w).Encode(status)
	if err != nil {
		logging.Error("failed to encode health status", logging.Err("error", err))
	}
}

// handleReady returns 200 OK if bot is connected (for Kubernetes readiness probe).
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.isBotConnected() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not Ready"))
	}
}

// handleLive returns 200 OK if bot is running (for Kubernetes liveness probe).
func (s *Server) handleLive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Bot is considered alive if the server is running
	// The server itself being up is sufficient for liveness
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// getHealthStatus builds the current health status.
func (s *Server) getHealthStatus() HealthStatus {
	var connected = s.isBotConnected()
	var status = "healthy"
	if !connected {
		status = "degraded"
	}

	return HealthStatus{
		Status:         status,
		BotConnected:   connected,
		CommandCount:   len(commands.CommandDefinitions),
		Uptime:         time.Since(s.startTime),
		Timestamp:      time.Now().UTC(),
	}
}

// isBotConnected checks if the Discord session is connected and ready.
func (s *Server) isBotConnected() bool {
	if s.bot == nil {
		return false
	}
	return s.bot.DataReady
}

// Start starts the health check server in a goroutine.
// Returns immediately; the server runs in the background.
func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	logging.Info("starting health check server", logging.String("address", addr))

	// Start server in a goroutine
	go func() {
		var err = s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logging.Error("health server error", logging.Err("error", err))
		}
	}()

	return nil
}

// Stop gracefully shuts down the health check server with the given context timeout.
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	logging.Info("shutting down health check server")

	var err = s.httpServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	logging.Info("health check server stopped")
	return nil
}
