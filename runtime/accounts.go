package runtime

import "github.com/MrBlindMouse/Portico/identity"

// AccountStore is an alias for the dev plaintext account backend.
// Prefer identity.OpenFileAccountStore for production sites.
type AccountStore = identity.PlaintextAccountStore

// NewAccountStore creates a dev-only plaintext account store.
func NewAccountStore() *AccountStore {
	return identity.NewPlaintextAccountStore()
}
