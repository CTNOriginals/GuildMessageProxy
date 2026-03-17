package events

import (
	"github.com/CTNOriginals/GuildMessageProxy/internal/logging"
	"github.com/bwmarrin/discordgo"
)

// ErrorCategory represents the classification of an error for appropriate handling.
type ErrorCategory string

const (
	// CategoryTransient represents retryable errors (429, 502, 503).
	CategoryTransient ErrorCategory = "transient"
	// CategoryPermanentAuth represents auth failures (40001, 40004, invalid token).
	CategoryPermanentAuth ErrorCategory = "permanent_auth"
	// CategoryPermanentResource represents missing resources (10003, 10008, 10013).
	CategoryPermanentResource ErrorCategory = "permanent_resource"
	// CategoryValidation represents invalid data (50035, 50016).
	CategoryValidation ErrorCategory = "validation"
)

// Discord error codes mapped to categories.
const (
	// Transient error codes
	discordErrRateLimit      = 429
	discordErrServerError    = 502
	discordErrServiceUnavail = 503

	// Permanent auth error codes
	discordErrUnauthorized       = 40001
	discordErrDisallowedIntent   = 40004

	// Permanent resource error codes
	discordErrUnknownChannel = 10003
	discordErrUnknownMessage = 10008
	discordErrUnknownUser    = 10013

	// Validation error codes
	discordErrInvalidFormBody = 50035
	discordErrTooManyAttachs  = 50016
)

// LogError logs an error with context information.
// Maintains backward compatibility with existing code.
func LogError(err error, context string) {
	if err == nil {
		return
	}
	category := CategorizeDiscordError(err)
	LogErrorWithCategory(err, category, context, 0)
}

// LogErrorWithCategory logs an error with categorization and retry context.
// Uses appropriate log level based on category.
func LogErrorWithCategory(err error, category ErrorCategory, context string, retryCount int) {
	if err == nil {
		return
	}

	switch category {
	case CategoryTransient:
		logging.Warn("Transient error, will retry",
			logging.Err("error", err),
			logging.String("category", string(category)),
			logging.Int("retry_count", retryCount),
			logging.Int("max_retries", 3),
			logging.Int("backoff_ms", calculateBackoff(retryCount)),
			logging.Int("discord_error_code", extractDiscordErrorCode(err)),
			logging.String("context", context),
		)
	case CategoryPermanentAuth:
		logging.Fatal("Authentication failed, shutting down",
			logging.Err("error", err),
			logging.Int("discord_error_code", extractDiscordErrorCode(err)),
			logging.String("context", context),
		)
	case CategoryPermanentResource:
		logging.Error("Resource not found",
			logging.Err("error", err),
			logging.String("category", string(category)),
			logging.Int("discord_error_code", extractDiscordErrorCode(err)),
			logging.String("resource_type", inferResourceType(err)),
			logging.String("context", context),
		)
	case CategoryValidation:
		logging.Warn("Validation failed",
			logging.Err("error", err),
			logging.String("category", string(category)),
			logging.Int("discord_error_code", extractDiscordErrorCode(err)),
			logging.String("context", context),
		)
	default:
		logging.Error("Operation failed",
			logging.Err("error", err),
			logging.String("context", context),
		)
	}
}

// CategorizeDiscordError extracts the Discord error code and maps it to a category.
func CategorizeDiscordError(err error) ErrorCategory {
	if err == nil {
		return ""
	}

	code := extractDiscordErrorCode(err)
	if code == 0 {
		return ""
	}

	switch code {
	case discordErrRateLimit, discordErrServerError, discordErrServiceUnavail:
		return CategoryTransient
	case discordErrUnauthorized, discordErrDisallowedIntent:
		return CategoryPermanentAuth
	case discordErrUnknownChannel, discordErrUnknownMessage, discordErrUnknownUser:
		return CategoryPermanentResource
	case discordErrInvalidFormBody, discordErrTooManyAttachs:
		return CategoryValidation
	default:
		return ""
	}
}

// extractDiscordErrorCode attempts to extract the Discord JSON error code from an error.
func extractDiscordErrorCode(err error) int {
	if err == nil {
		return 0
	}

	// Check if it's a RESTError from discordgo
	if restErr, ok := err.(*discordgo.RESTError); ok && restErr.Message != nil {
		return restErr.Message.Code
	}

	return 0
}

// inferResourceType attempts to determine the resource type from error context.
func inferResourceType(err error) string {
	if err == nil {
		return "unknown"
	}

	code := extractDiscordErrorCode(err)
	switch code {
	case discordErrUnknownChannel:
		return "channel"
	case discordErrUnknownMessage:
		return "message"
	case discordErrUnknownUser:
		return "user"
	default:
		return "unknown"
	}
}

// calculateBackoff calculates exponential backoff in milliseconds.
func calculateBackoff(retryCount int) int {
	if retryCount <= 0 {
		return 0
	}
	// Exponential backoff: 1000ms, 2000ms, 4000ms
	return 1000 * (1 << (retryCount - 1))
}

// RespondToUser sends an ephemeral message to the user who triggered an interaction.
// Use this to provide feedback when something goes wrong.
func RespondToUser(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	var err error = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		logging.Error("Failed to respond to user",
			logging.Err("error", err),
			logging.String("user_id", i.Member.User.ID),
			logging.String("guild_id", i.GuildID),
		)
	}
}

// RespondWithError sends a formatted error response to the user.
// Includes both a user-friendly message and logs the actual error.
func RespondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, userMsg string, err error) {
	if err != nil {
		// Log the error with context including Discord interaction details
		logging.Error("Error responding to user",
			logging.Err("error", err),
			logging.String("context", userMsg),
			logging.String("user_id", i.Member.User.ID),
			logging.String("guild_id", i.GuildID),
			logging.String("interaction_id", i.ID),
		)
		// Also use the existing LogError for backward compatibility and categorization
		LogError(err, userMsg)
	}
	RespondToUser(s, i, userMsg)
}
