package events

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// mockRESTError creates a discordgo.RESTError with the specified code
func mockRESTError(code int) *discordgo.RESTError {
	return &discordgo.RESTError{
		Message: &discordgo.APIErrorMessage{
			Code:    code,
			Message: "test error message",
		},
	}
}

// TestExtractDiscordErrorCode_WithRESTError tests extraction from discordgo.RESTError
func TestExtractDiscordErrorCode_WithRESTError(t *testing.T) {
	var tests = []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "rate limit error 429",
			err:      mockRESTError(429),
			expected: 429,
		},
		{
			name:     "server error 502",
			err:      mockRESTError(502),
			expected: 502,
		},
		{
			name:     "service unavailable 503",
			err:      mockRESTError(503),
			expected: 503,
		},
		{
			name:     "unauthorized 40001",
			err:      mockRESTError(40001),
			expected: 40001,
		},
		{
			name:     "disallowed intent 40004",
			err:      mockRESTError(40004),
			expected: 40004,
		},
		{
			name:     "unknown channel 10003",
			err:      mockRESTError(10003),
			expected: 10003,
		},
		{
			name:     "unknown message 10008",
			err:      mockRESTError(10008),
			expected: 10008,
		},
		{
			name:     "unknown user 10013",
			err:      mockRESTError(10013),
			expected: 10013,
		},
		{
			name:     "invalid form body 50035",
			err:      mockRESTError(50035),
			expected: 50035,
		},
		{
			name:     "too many attachments 50016",
			err:      mockRESTError(50016),
			expected: 50016,
		},
		{
			name:     "nil error returns 0",
			err:      nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result int = extractDiscordErrorCode(tt.err)
			if result != tt.expected {
				t.Errorf("extractDiscordErrorCode() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestExtractDiscordErrorCode_WithNonRESTError tests extraction with non-RESTError types
func TestExtractDiscordErrorCode_WithNonRESTError(t *testing.T) {
	var genericErr error = errors.New("generic error")
	var result int = extractDiscordErrorCode(genericErr)

	if result != 0 {
		t.Errorf("extractDiscordErrorCode() with generic error = %d, want 0", result)
	}
}

// TestExtractDiscordErrorCode_WithRESTErrorNilMessage tests RESTError with nil Message
func TestExtractDiscordErrorCode_WithRESTErrorNilMessage(t *testing.T) {
	var restErr *discordgo.RESTError = &discordgo.RESTError{
		Message: nil,
	}
	var result int = extractDiscordErrorCode(restErr)

	if result != 0 {
		t.Errorf("extractDiscordErrorCode() with nil Message = %d, want 0", result)
	}
}

// TestCalculateBackoff tests the exponential backoff calculation
func TestCalculateBackoff(t *testing.T) {
	var tests = []struct {
		name       string
		retryCount int
		expected   int
	}{
		{
			name:       "zero retries returns 0",
			retryCount: 0,
			expected:   0,
		},
		{
			name:       "negative retries returns 0",
			retryCount: -1,
			expected:   0,
		},
		{
			name:       "first retry returns 1000ms",
			retryCount: 1,
			expected:   1000,
		},
		{
			name:       "second retry returns 2000ms",
			retryCount: 2,
			expected:   2000,
		},
		{
			name:       "third retry returns 4000ms",
			retryCount: 3,
			expected:   4000,
		},
		{
			name:       "fourth retry returns 8000ms",
			retryCount: 4,
			expected:   8000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result int = calculateBackoff(tt.retryCount)
			if result != tt.expected {
				t.Errorf("calculateBackoff(%d) = %d, want %d", tt.retryCount, result, tt.expected)
			}
		})
	}
}

// TestCategorizeDiscordError_TransientErrors tests transient error categorization
func TestCategorizeDiscordError_TransientErrors(t *testing.T) {
	var tests = []struct {
		name     string
		code     int
		expected ErrorCategory
	}{
		{
			name:     "rate limit 429 is transient",
			code:     429,
			expected: CategoryTransient,
		},
		{
			name:     "server error 502 is transient",
			code:     502,
			expected: CategoryTransient,
		},
		{
			name:     "service unavailable 503 is transient",
			code:     503,
			expected: CategoryTransient,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error = mockRESTError(tt.code)
			var result ErrorCategory = CategorizeDiscordError(err)
			if result != tt.expected {
				t.Errorf("CategorizeDiscordError() with code %d = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

// TestCategorizeDiscordError_PermanentAuthErrors tests permanent auth error categorization
func TestCategorizeDiscordError_PermanentAuthErrors(t *testing.T) {
	var tests = []struct {
		name     string
		code     int
		expected ErrorCategory
	}{
		{
			name:     "unauthorized 40001 is permanent auth",
			code:     40001,
			expected: CategoryPermanentAuth,
		},
		{
			name:     "disallowed intent 40004 is permanent auth",
			code:     40004,
			expected: CategoryPermanentAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error = mockRESTError(tt.code)
			var result ErrorCategory = CategorizeDiscordError(err)
			if result != tt.expected {
				t.Errorf("CategorizeDiscordError() with code %d = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

// TestCategorizeDiscordError_PermanentResourceErrors tests permanent resource error categorization
func TestCategorizeDiscordError_PermanentResourceErrors(t *testing.T) {
	var tests = []struct {
		name     string
		code     int
		expected ErrorCategory
	}{
		{
			name:     "unknown channel 10003 is permanent resource",
			code:     10003,
			expected: CategoryPermanentResource,
		},
		{
			name:     "unknown message 10008 is permanent resource",
			code:     10008,
			expected: CategoryPermanentResource,
		},
		{
			name:     "unknown user 10013 is permanent resource",
			code:     10013,
			expected: CategoryPermanentResource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error = mockRESTError(tt.code)
			var result ErrorCategory = CategorizeDiscordError(err)
			if result != tt.expected {
				t.Errorf("CategorizeDiscordError() with code %d = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

// TestCategorizeDiscordError_ValidationErrors tests validation error categorization
func TestCategorizeDiscordError_ValidationErrors(t *testing.T) {
	var tests = []struct {
		name     string
		code     int
		expected ErrorCategory
	}{
		{
			name:     "invalid form body 50035 is validation",
			code:     50035,
			expected: CategoryValidation,
		},
		{
			name:     "too many attachments 50016 is validation",
			code:     50016,
			expected: CategoryValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error = mockRESTError(tt.code)
			var result ErrorCategory = CategorizeDiscordError(err)
			if result != tt.expected {
				t.Errorf("CategorizeDiscordError() with code %d = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

// TestCategorizeDiscordError_EdgeCases tests edge cases
func TestCategorizeDiscordError_EdgeCases(t *testing.T) {
	var tests = []struct {
		name     string
		err      error
		expected ErrorCategory
	}{
		{
			name:     "nil error returns empty category",
			err:      nil,
			expected: "",
		},
		{
			name:     "generic error returns empty category",
			err:      errors.New("generic error"),
			expected: "",
		},
		{
			name:     "unknown error code returns empty category",
			err:      mockRESTError(99999),
			expected: "",
		},
		{
			name:     "RESTError with nil message returns empty category",
			err:      &discordgo.RESTError{Message: nil},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ErrorCategory = CategorizeDiscordError(tt.err)
			if result != tt.expected {
				t.Errorf("CategorizeDiscordError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestInferResourceType tests resource type inference
func TestInferResourceType(t *testing.T) {
	var tests = []struct {
		name     string
		code     int
		expected string
	}{
		{
			name:     "unknown channel returns channel",
			code:     10003,
			expected: "channel",
		},
		{
			name:     "unknown message returns message",
			code:     10008,
			expected: "message",
		},
		{
			name:     "unknown user returns user",
			code:     10013,
			expected: "user",
		},
		{
			name:     "unknown code returns unknown",
			code:     99999,
			expected: "unknown",
		},
		{
			name:     "nil error returns unknown",
			code:     -1, // sentinel for nil
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.code == -1 {
				err = nil
			} else {
				err = mockRESTError(tt.code)
			}
			var result string = inferResourceType(err)
			if result != tt.expected {
				t.Errorf("inferResourceType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestErrorCategoryConstants tests that error category constants are defined correctly
func TestErrorCategoryConstants(t *testing.T) {
	var tests = []struct {
		name     string
		category ErrorCategory
		expected string
	}{
		{
			name:     "CategoryTransient is transient",
			category: CategoryTransient,
			expected: "transient",
		},
		{
			name:     "CategoryPermanentAuth is permanent_auth",
			category: CategoryPermanentAuth,
			expected: "permanent_auth",
		},
		{
			name:     "CategoryPermanentResource is permanent_resource",
			category: CategoryPermanentResource,
			expected: "permanent_resource",
		},
		{
			name:     "CategoryValidation is validation",
			category: CategoryValidation,
			expected: "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.category) != tt.expected {
				t.Errorf("ErrorCategory constant = %v, want %v", string(tt.category), tt.expected)
			}
		})
	}
}

// TestDiscordErrorCodeConstants tests that error code constants are defined correctly
func TestDiscordErrorCodeConstants(t *testing.T) {
	// Transient errors
	if discordErrRateLimit != 429 {
		t.Errorf("discordErrRateLimit = %d, want 429", discordErrRateLimit)
	}
	if discordErrServerError != 502 {
		t.Errorf("discordErrServerError = %d, want 502", discordErrServerError)
	}
	if discordErrServiceUnavail != 503 {
		t.Errorf("discordErrServiceUnavail = %d, want 503", discordErrServiceUnavail)
	}

	// Permanent auth errors
	if discordErrUnauthorized != 40001 {
		t.Errorf("discordErrUnauthorized = %d, want 40001", discordErrUnauthorized)
	}
	if discordErrDisallowedIntent != 40004 {
		t.Errorf("discordErrDisallowedIntent = %d, want 40004", discordErrDisallowedIntent)
	}

	// Permanent resource errors
	if discordErrUnknownChannel != 10003 {
		t.Errorf("discordErrUnknownChannel = %d, want 10003", discordErrUnknownChannel)
	}
	if discordErrUnknownMessage != 10008 {
		t.Errorf("discordErrUnknownMessage = %d, want 10008", discordErrUnknownMessage)
	}
	if discordErrUnknownUser != 10013 {
		t.Errorf("discordErrUnknownUser = %d, want 10013", discordErrUnknownUser)
	}

	// Validation errors
	if discordErrInvalidFormBody != 50035 {
		t.Errorf("discordErrInvalidFormBody = %d, want 50035", discordErrInvalidFormBody)
	}
	if discordErrTooManyAttachs != 50016 {
		t.Errorf("discordErrTooManyAttachs = %d, want 50016", discordErrTooManyAttachs)
	}
}
