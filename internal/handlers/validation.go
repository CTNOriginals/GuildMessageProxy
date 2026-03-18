package handlers

import (
	"fmt"

	"github.com/CTNOriginals/GuildMessageProxy/internal/storage"
)

// ValidationResult holds validation outcome
type ValidationResult struct {
	Valid bool
	Error string
}

// MaxMessageLength is Discord's message character limit
const MaxMessageLength int = 2000

// ValidateContent checks if message content is valid for posting.
// Returns ValidationResult with Valid=true if content passes all checks.
func ValidateContent(content string) ValidationResult {
	// Check empty content
	if content == "" {
		return ValidationResult{
			Valid: false,
			Error: "Message content cannot be empty.",
		}
	}

	// Check max length (2000 chars for Discord)
	if len(content) > MaxMessageLength {
		return ValidationResult{
			Valid: false,
			Error: fmt.Sprintf("Message exceeds maximum length of %d characters (current: %d).", MaxMessageLength, len(content)),
		}
	}

	return ValidationResult{
		Valid: true,
		Error: "",
	}
}

// ValidateEditPermission checks if a user can edit a proxy message.
// MVP: Only the original owner can edit.
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
