package health

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

// mockSession is a minimal mock of discordgo.Session for testing.
type mockSession struct {
	dataReady bool
}

func (m *mockSession) StateUser() *discordgo.User {
	return nil
}

// TestHealthStatus_Struct tests the HealthStatus struct JSON marshaling.
func TestHealthStatus_Struct(t *testing.T) {
	var status = HealthStatus{
		Status:         "healthy",
		BotConnected:   true,
		CommandCount:   5,
		Uptime:         time.Hour,
		Timestamp:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	var data, err = json.Marshal(status)
	if err != nil {
		t.Fatalf("Failed to marshal HealthStatus: %v", err)
	}

	var jsonStr = string(data)

	if !strings.Contains(jsonStr, `"status":"healthy"`) {
		t.Errorf("JSON missing status field: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"bot_connected":true`) {
		t.Errorf("JSON missing bot_connected field: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"command_count":5`) {
		t.Errorf("JSON missing command_count field: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"uptime":`) {
		t.Errorf("JSON missing uptime field: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"timestamp":`) {
		t.Errorf("JSON missing timestamp field: %s", jsonStr)
	}
}

// TestNewServer tests the server creation.
func TestNewServer(t *testing.T) {
	var startTime = time.Now()
	var server = NewServer(nil, startTime)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.bot != nil {
		t.Error("Expected bot to be nil")
	}

	if !server.startTime.Equal(startTime) {
		t.Error("Expected startTime to match")
	}

	if server.mux == nil {
		t.Error("Expected mux to be initialized")
	}
}

// TestServer_StartStop tests starting and stopping the server.
func TestServer_StartStop(t *testing.T) {
	var startTime = time.Now()
	var server = NewServer(nil, startTime)

	var err = server.Start("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
}

// TestHandleHealth tests the /health endpoint.
func TestHandleHealth(t *testing.T) {
	var startTime = time.Now()
	var server = NewServer(nil, startTime)

	var err = server.Start("127.0.0.1:18080")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	defer func() {
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	var resp, err2 = http.Get("http://127.0.0.1:18080/health")
	if err2 != nil {
		t.Fatalf("Failed to make request: %v", err2)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var contentType = resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var body, err3 = io.ReadAll(resp.Body)
	if err3 != nil {
		t.Fatalf("Failed to read body: %v", err3)
	}

	var status HealthStatus
	err = json.Unmarshal(body, &status)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if status.Status != "degraded" {
		t.Errorf("Expected status 'degraded' (bot not connected), got '%s'", status.Status)
	}

	if status.BotConnected {
		t.Error("Expected BotConnected to be false")
	}

	if status.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

// TestHandleHealth_MethodNotAllowed tests that non-GET methods are rejected.
func TestHandleHealth_MethodNotAllowed(t *testing.T) {
	var startTime = time.Now()
	var server = NewServer(nil, startTime)

	var err = server.Start("127.0.0.1:18081")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	defer func() {
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	var resp, err2 = http.Post("http://127.0.0.1:18081/health", "application/json", nil)
	if err2 != nil {
		t.Fatalf("Failed to make request: %v", err2)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

// TestHandleReady_NotConnected tests /ready when bot is not connected.
func TestHandleReady_NotConnected(t *testing.T) {
	var startTime = time.Now()
	var server = NewServer(nil, startTime)

	var err = server.Start("127.0.0.1:18082")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	defer func() {
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	var resp, err2 = http.Get("http://127.0.0.1:18082/ready")
	if err2 != nil {
		t.Fatalf("Failed to make request: %v", err2)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}
}

// TestHandleLive tests the /live endpoint.
func TestHandleLive(t *testing.T) {
	var startTime = time.Now()
	var server = NewServer(nil, startTime)

	var err = server.Start("127.0.0.1:18083")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	defer func() {
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	var resp, err2 = http.Get("http://127.0.0.1:18083/live")
	if err2 != nil {
		t.Fatalf("Failed to make request: %v", err2)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var body, err3 = io.ReadAll(resp.Body)
	if err3 != nil {
		t.Fatalf("Failed to read body: %v", err3)
	}

	if string(body) != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", string(body))
	}
}

// TestIsBotConnected_NilSession tests isBotConnected with nil session.
func TestIsBotConnected_NilSession(t *testing.T) {
	var startTime = time.Now()
	var server = NewServer(nil, startTime)

	if server.isBotConnected() {
		t.Error("Expected isBotConnected to return false for nil session")
	}
}

// TestGetHealthStatus tests the health status generation.
func TestGetHealthStatus(t *testing.T) {
	var startTime = time.Now().Add(-time.Hour)
	var server = NewServer(nil, startTime)

	var status = server.getHealthStatus()

	if status.Status != "degraded" {
		t.Errorf("Expected status 'degraded', got '%s'", status.Status)
	}

	if status.BotConnected {
		t.Error("Expected BotConnected to be false")
	}

	if status.Uptime < time.Hour {
		t.Errorf("Expected Uptime >= 1 hour, got %v", status.Uptime)
	}

	if status.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}
