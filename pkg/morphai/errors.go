package morphai

import (
	"errors"
	"fmt"
)

// ErrProviderNotConfigured means the selected provider has no credentials
// (or, for a provider that does not use a key, no base URL). A settings UI
// can errors.Is this and ask the user to add a key. The client does not
// substitute a different provider.
var ErrProviderNotConfigured = errors.New("ai provider is not configured")

// ErrCapabilityUnsupported means the selected provider cannot perform the
// requested operation (streaming, tools, vision, or JSON mode).
var ErrCapabilityUnsupported = errors.New("ai provider does not support this capability")

// ErrUnknownProvider means Provider is not in the registry.
var ErrUnknownProvider = errors.New("unknown ai provider")

// ProviderConfigError is returned when a named provider cannot be called.
// It unwraps to ErrProviderNotConfigured.
type ProviderConfigError struct {
	Provider ProviderID
	Detail   string
}

func (e *ProviderConfigError) Error() string {
	if e == nil {
		return ErrProviderNotConfigured.Error()
	}
	if e.Detail == "" {
		return fmt.Sprintf("provider %s is not configured", e.Provider)
	}
	return fmt.Sprintf("provider %s is not configured: %s", e.Provider, e.Detail)
}

func (e *ProviderConfigError) Unwrap() error { return ErrProviderNotConfigured }

// UnknownProviderError unwraps to ErrUnknownProvider.
type UnknownProviderError struct {
	ID ProviderID
}

func (e *UnknownProviderError) Error() string {
	if e == nil {
		return ErrUnknownProvider.Error()
	}
	return fmt.Sprintf("unknown ai provider %q", e.ID)
}

func (e *UnknownProviderError) Unwrap() error { return ErrUnknownProvider }

// CapabilityError unwraps to ErrCapabilityUnsupported.
type CapabilityError struct {
	Provider   ProviderID
	Capability string
	// Detail, when set, is the full error text. The legacy native-vision
	// message is preserved this way.
	Detail string
}

func (e *CapabilityError) Error() string {
	if e == nil {
		return ErrCapabilityUnsupported.Error()
	}
	if e.Detail != "" {
		return e.Detail
	}
	name := string(e.Provider)
	if name == "" {
		name = "provider"
	}
	return fmt.Sprintf("provider %s does not support %s", name, e.Capability)
}

func (e *CapabilityError) Unwrap() error { return ErrCapabilityUnsupported }

// APIError is a non-2xx response from a provider, with the provider's code
// and message when the body could be parsed.
type APIError struct {
	Provider   ProviderID
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e == nil {
		return "API error"
	}
	return fmt.Sprintf("API error (status %d): %s - %s", e.StatusCode, e.Code, e.Message)
}

// legacyKeyError is the historical missing-key error for callers that never
// select a provider. The text stays exact; errors.Is still matches
// ErrProviderNotConfigured.
type legacyKeyError struct{}

func (legacyKeyError) Error() string { return "MORPH_AI_API_KEY is not configured" }

func (legacyKeyError) Unwrap() error { return ErrProviderNotConfigured }
