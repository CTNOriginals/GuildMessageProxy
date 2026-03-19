package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/CTNOriginals/GuildMessageProxy/internal/events"
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/bwmarrin/discordgo"
)

// Common errors.
var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrMaxRetriesReached  = errors.New("max retries reached")
	ErrContextCancelled   = errors.New("context cancelled")
)

// RateLimiter provides token bucket rate limiting for API calls.
type RateLimiter interface {
	Wait(ctx context.Context) error
}

// TokenBucket implements a thread-safe token bucket rate limiter.
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	rate       float64 // tokens per second
	lastRefill time.Time
}

// NewTokenBucket creates a new token bucket rate limiter.
// rate: tokens per second (requests per second)
// burst: maximum tokens that can be accumulated (burst capacity)
func NewTokenBucket(rate, burst float64) *TokenBucket {
	var tb TokenBucket = TokenBucket{
		tokens:     burst,
		maxTokens:  burst,
		rate:       rate,
		lastRefill: time.Now(),
	}
	return &tb
}

// Wait blocks until a token is available or context is cancelled.
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if err := tb.tryAcquire(); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ErrContextCancelled
		case <-time.After(10 * time.Millisecond):
			// Continue and try again
		}
	}
}

// tryAcquire attempts to acquire a token without blocking.
func (tb *TokenBucket) tryAcquire() error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	var now time.Time = time.Now()
	var elapsed time.Duration = now.Sub(tb.lastRefill)
	tb.lastRefill = now

	// Refill tokens based on elapsed time
	tb.tokens += elapsed.Seconds() * tb.rate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	if tb.tokens >= 1 {
		tb.tokens--
		return nil
	}

	return ErrRateLimitExceeded
}

// RetryConfig configures exponential backoff retry behavior.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns the standard retry configuration.
// Max 3 retries with exponential backoff starting at 1s, max 30s.
func DefaultRetryConfig() RetryConfig {
	var cfg RetryConfig = RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
	}
	return cfg
}

// calculateDelay computes the backoff delay for a given retry attempt.
func (rc RetryConfig) calculateDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	// Exponential backoff: baseDelay * 2^(attempt-1)
	var delay time.Duration = rc.BaseDelay * time.Duration(1<<(attempt-1))

	if delay > rc.MaxDelay {
		delay = rc.MaxDelay
	}

	return delay
}

// State represents the circuit breaker state.
type State int

const (
	// StateClosed allows requests through normally.
	StateClosed State = iota
	// StateOpen blocks all requests due to consecutive failures.
	StateOpen
	// StateHalfOpen allows a limited number of requests to test recovery.
	StateHalfOpen
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// CircuitBreaker prevents repeated calls to failing services.
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	failureThreshold int
	successThreshold int
	cooldownDuration time.Duration
}

// NewCircuitBreaker creates a new circuit breaker with default settings.
// Opens after 5 consecutive failures.
// Half-open after 30s cooldown.
// Closes after 2 consecutive successes in half-open state.
func NewCircuitBreaker() *CircuitBreaker {
	var cb CircuitBreaker = CircuitBreaker{
		state:            StateClosed,
		failureThreshold: 5,
		successThreshold: 2,
		cooldownDuration: 30 * time.Second,
	}
	return &cb
}

// Allow returns nil if the request should proceed, or an error if blocked.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.transitionState()

	if cb.state == StateOpen {
		return ErrCircuitBreakerOpen
	}

	return nil
}

// RecordSuccess records a successful operation and updates state.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			logging.Info("Circuit breaker closed after recovery",
				logging.String("previous_state", cb.state.String()),
				logging.Int("success_count", cb.successCount),
			)
			cb.state = StateClosed
			cb.successCount = 0
		}
	} else {
		cb.successCount = 0
	}
}

// RecordFailure records a failed operation and updates state.
func (cb *CircuitBreaker) RecordFailure(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen {
		// Any failure in half-open immediately reopens
		logging.Warn("Circuit breaker reopened due to failure in half-open state",
			logging.String("error", err.Error()),
		)
		cb.state = StateOpen
		cb.successCount = 0
	} else if cb.state == StateClosed && cb.failureCount >= cb.failureThreshold {
		logging.Warn("Circuit breaker opened due to consecutive failures",
			logging.Int("failure_count", cb.failureCount),
			logging.String("error", err.Error()),
		)
		cb.state = StateOpen
	}
}

// transitionState handles state transitions based on time and current state.
func (cb *CircuitBreaker) transitionState() {
	if cb.state == StateOpen {
		var elapsed time.Duration = time.Since(cb.lastFailureTime)
		if elapsed >= cb.cooldownDuration {
			logging.Info("Circuit breaker entering half-open state",
				logging.Duration("cooldown_duration", elapsed),
			)
			cb.state = StateHalfOpen
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// DiscordClient wraps discordgo.Session with rate limiting, retry logic, and circuit breaker.
type DiscordClient struct {
	session        *discordgo.Session
	rateLimiter    RateLimiter
	circuitBreaker *CircuitBreaker
	retryConfig    RetryConfig
}

// NewDiscordClient creates a new wrapped Discord client with reliability features.
func NewDiscordClient(session *discordgo.Session) *DiscordClient {
	var client DiscordClient = DiscordClient{
		session:        session,
		rateLimiter:    NewTokenBucket(5, 10), // 5 req/s, burst 10
		circuitBreaker: NewCircuitBreaker(),
		retryConfig:    DefaultRetryConfig(),
	}
	return &client
}

// NewDiscordClientWithConfig creates a client with custom configuration.
func NewDiscordClientWithConfig(
	session *discordgo.Session,
	rateLimiter RateLimiter,
	circuitBreaker *CircuitBreaker,
	retryConfig RetryConfig,
) *DiscordClient {
	var client DiscordClient = DiscordClient{
		session:        session,
		rateLimiter:    rateLimiter,
		circuitBreaker: circuitBreaker,
		retryConfig:    retryConfig,
	}
	return &client
}

// executeWithRetry executes a Discord API call with rate limiting, circuit breaker, and retry logic.
func (dc *DiscordClient) executeWithRetry(ctx context.Context, operation string, fn func() error) error {
	// Check circuit breaker first
	if err := dc.circuitBreaker.Allow(); err != nil {
		logging.Warn("Circuit breaker blocked request",
			logging.String("operation", operation),
			logging.String("state", dc.circuitBreaker.State().String()),
		)
		return err
	}

	// Wait for rate limiter
	if err := dc.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	// Execute with retry logic
	var lastErr error
	for attempt := 0; attempt <= dc.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			var delay time.Duration = dc.retryConfig.calculateDelay(attempt)
			logging.Warn("Retrying after transient error",
				logging.String("operation", operation),
				logging.Int("attempt", attempt),
				logging.Int("max_retries", dc.retryConfig.MaxRetries),
				logging.Duration("delay", delay),
			)

			select {
			case <-ctx.Done():
				return ErrContextCancelled
			case <-time.After(delay):
				// Continue to retry
			}

			// Wait for rate limiter before retry
			if err := dc.rateLimiter.Wait(ctx); err != nil {
				return err
			}
		}

		lastErr = fn()

		if lastErr == nil {
			dc.circuitBreaker.RecordSuccess()
			return nil
		}

		// Check if error is transient and worth retrying
		var category events.ErrorCategory = events.CategorizeDiscordError(lastErr)
		if category != events.CategoryTransient {
			// Non-transient error, fail immediately
			dc.circuitBreaker.RecordFailure(lastErr)
			return lastErr
		}

		// Log transient error for retry
		events.LogErrorWithCategory(lastErr, category, operation, attempt)
	}

	// All retries exhausted
	dc.circuitBreaker.RecordFailure(lastErr)
	return fmt.Errorf("%w: %v", ErrMaxRetriesReached, lastErr)
}

// ChannelMessageSend sends a message to a channel with reliability features.
func (dc *DiscordClient) ChannelMessageSend(ctx context.Context, channelID, content string) (*discordgo.Message, error) {
	var msg *discordgo.Message
	var err error = dc.executeWithRetry(ctx, "ChannelMessageSend", func() error {
		var sendErr error
		msg, sendErr = dc.session.ChannelMessageSend(channelID, content)
		return sendErr
	})
	return msg, err
}

// ChannelMessageSendComplex sends a complex message to a channel with reliability features.
func (dc *DiscordClient) ChannelMessageSendComplex(ctx context.Context, channelID string, data *discordgo.MessageSend) (*discordgo.Message, error) {
	var msg *discordgo.Message
	var err error = dc.executeWithRetry(ctx, "ChannelMessageSendComplex", func() error {
		var sendErr error
		msg, sendErr = dc.session.ChannelMessageSendComplex(channelID, data)
		return sendErr
	})
	return msg, err
}

// ChannelMessageEdit edits a message with reliability features.
func (dc *DiscordClient) ChannelMessageEdit(ctx context.Context, channelID, messageID, content string) (*discordgo.Message, error) {
	var msg *discordgo.Message
	var err error = dc.executeWithRetry(ctx, "ChannelMessageEdit", func() error {
		var editErr error
		msg, editErr = dc.session.ChannelMessageEdit(channelID, messageID, content)
		return editErr
	})
	return msg, err
}

// ChannelMessageDelete deletes a message with reliability features.
func (dc *DiscordClient) ChannelMessageDelete(ctx context.Context, channelID, messageID string) error {
	var err error = dc.executeWithRetry(ctx, "ChannelMessageDelete", func() error {
		return dc.session.ChannelMessageDelete(channelID, messageID)
	})
	return err
}

// InteractionRespond responds to an interaction with reliability features.
func (dc *DiscordClient) InteractionRespond(ctx context.Context, interaction *discordgo.Interaction, resp *discordgo.InteractionResponse) error {
	var err error = dc.executeWithRetry(ctx, "InteractionRespond", func() error {
		return dc.session.InteractionRespond(interaction, resp)
	})
	return err
}

// Guild retrieves guild information with reliability features.
func (dc *DiscordClient) Guild(ctx context.Context, guildID string) (*discordgo.Guild, error) {
	var guild *discordgo.Guild
	var err error = dc.executeWithRetry(ctx, "Guild", func() error {
		var guildErr error
		guild, guildErr = dc.session.Guild(guildID)
		return guildErr
	})
	return guild, err
}

// Channel retrieves channel information with reliability features.
func (dc *DiscordClient) Channel(ctx context.Context, channelID string) (*discordgo.Channel, error) {
	var channel *discordgo.Channel
	var err error = dc.executeWithRetry(ctx, "Channel", func() error {
		var chErr error
		channel, chErr = dc.session.Channel(channelID)
		return chErr
	})
	return channel, err
}

// User retrieves user information with reliability features.
func (dc *DiscordClient) User(ctx context.Context, userID string) (*discordgo.User, error) {
	var user *discordgo.User
	var err error = dc.executeWithRetry(ctx, "User", func() error {
		var userErr error
		user, userErr = dc.session.User(userID)
		return userErr
	})
	return user, err
}

// Session returns the underlying discordgo session.
func (dc *DiscordClient) Session() *discordgo.Session {
	return dc.session
}

// RateLimiter returns the rate limiter for inspection or testing.
func (dc *DiscordClient) RateLimiter() RateLimiter {
	return dc.rateLimiter
}

// CircuitBreaker returns the circuit breaker for inspection or testing.
func (dc *DiscordClient) CircuitBreaker() *CircuitBreaker {
	return dc.circuitBreaker
}
