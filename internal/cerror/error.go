package cerror

import "fmt"

type Error struct {
	Operation string
	Cause     string
}

// Errors of the storage package
var (
	// If accessing the database fails
	ErrAccessDatabase = &Error{Operation: "Access database"}
	// If initializing or migrating the database fails
	ErrInitializeDatabase = &Error{Operation: "Initialize database"}
)

// Errors of the base package
var (
	// If initializing the base fails
	ErrInitializeBase = &Error{Operation: "Initialize base"}
	// If getting the master key from base fails
	ErrGetMaster = &Error{Operation: "Get master"}
	// If authentication with master key fails
	ErrAuthenticate = &Error{Operation: "Authenticate"}
)

// Errors of the secret package
var (
	ErrCreateSecret = &Error{Operation: "Create secret"}
	ErrListSecrets  = &Error{Operation: "List secrets"}
	ErrGetSecret    = &Error{Operation: "Get secret"}
	ErrDeleteSecret = &Error{Operation: "Delete secret"}
)

// Errors of the crypto package
var (
	ErrGetTerminalState = &Error{Operation: "Get terminal state"}
	ErrReadPassword     = &Error{Operation: "Read password"}
	ErrEncrypt          = &Error{Operation: "Encrypt"}
	ErrDecrypt          = &Error{Operation: "Decrypt"}
)

// Errors of the prompt package
var (
	ErrScanKey  = &Error{Operation: "Scan key", Cause: "key could not scanned"}
	ErrEmptyKey = &Error{Operation: "Read key", Cause: "key is empty"}
)

func (e *Error) Error() string {
	return fmt.Sprintf("Operation \"%s\" has failed: %s", e.Operation, e.Cause)
}
