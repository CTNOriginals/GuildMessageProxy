package handlers

import (
	"fmt"
	"strings"

	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// ValidationResult holds validation outcome.
type ValidationResult struct {
	// Valid is true if the validation passed and the operation is allowed.
	Valid bool
	// Error is a human-readable error message when Valid is false.
	Error string
}

// MaxMessageLength is Discord's message character limit constant.
// Discord allows up to 2000 characters per message.
const MaxMessageLength int = 2000

// ValidateContent checks if message content is valid for posting.
//
// Parameters:
//   - content: The message text to validate.
//
// Returns a ValidationResult with Valid=true if content passes all checks.
// If validation fails, Valid is false and Error contains the reason.
//
// Validation checks performed:
//   - Empty or whitespace-only content is rejected.
//   - Content exceeding MaxMessageLength (2000 chars) is rejected.
//
// Example error messages:
//   - "Please enter a message. Empty messages cannot be sent."
//   - "Your message is too long. Discord has a 2000 character limit..."
func ValidateContent(content string) ValidationResult {
	// Check empty or whitespace-only content
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ValidationResult{
			Valid: false,
			Error: "Please enter a message. Empty messages cannot be sent.",
		}
	}

	// Check max length (2000 chars for Discord)
	if len(content) > MaxMessageLength {
		return ValidationResult{
			Valid: false,
			Error: fmt.Sprintf("Your message is too long. Discord has a 2000 character limit, and your message is %d characters. Please shorten it.", len(content)),
		}
	}

	return ValidationResult{
		Valid: true,
		Error: "",
	}
}

// ValidateEditPermission checks if a user can edit a proxy message.
//
// Parameters:
//   - proxyMsg: The proxy message being targeted for editing.
//   - userID: The Discord ID of the user attempting to edit.
//
// Returns a ValidationResult with Valid=true if the user has permission.
// If validation fails, Valid is false and Error contains the reason.
//
// Current permission model (MVP):
//   - Only the original owner (message author) can edit their proxy messages.
//   - Nil proxyMsg is treated as "not found or no longer exists."
func ValidateEditPermission(proxyMsg *storage.ProxyMessage, userID string) ValidationResult {
	// Check if proxyMsg is nil
	if proxyMsg == nil {
		return ValidationResult{
			Valid: false,
			Error: "Message not found or no longer exists.",
		}
	}

	// Check if userID matches proxyMsg.OwnerID
	if proxyMsg.OwnerID != userID {
		return ValidationResult{
			Valid: false,
			Error: "Only the original message author can edit this message.",
		}
	}

	return ValidationResult{
		Valid: true,
		Error: "",
	}
}
