package secretbox

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// ScopeWorkspace is the associated-data scope for admin and workspace provider keys.
	ScopeWorkspace = "workspace"
	// ScopeUser is the associated-data scope for per-user provider keys.
	ScopeUser = "user"
)

// ErrInvalidAAD is returned when provider-key associated data fields are ambiguous or empty.
var ErrInvalidAAD = errors.New("secretbox: invalid associated data")

// ProviderKeyAAD returns the canonical associated data for a provider-key row:
// provider_key:<scope>:<ownerID>:<provider>.
// Scope is workspace or user. Fields must be non-empty and must not contain ':'.
func ProviderKeyAAD(scope, ownerID, provider string) ([]byte, error) {
	if scope != ScopeWorkspace && scope != ScopeUser {
		return nil, fmt.Errorf("%w: scope must be workspace or user", ErrInvalidAAD)
	}
	if !plainField(ownerID) || !plainField(provider) {
		return nil, fmt.Errorf("%w: owner and provider must be non-empty and must not contain a colon", ErrInvalidAAD)
	}
	return []byte("provider_key:" + scope + ":" + ownerID + ":" + provider), nil
}

func plainField(v string) bool {
	return v != "" && !strings.Contains(v, ":")
}
