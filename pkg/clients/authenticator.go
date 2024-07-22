package clients

// Authenticator is an interface for handling authentication with a service
type Authenticator interface {
	GetToken() string
	RefreshToken() error
	NeedsTokenRefresh() error
}
