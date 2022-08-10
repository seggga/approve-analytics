package models

// TokenPair represents JWT token
type TokenPair struct {
	Access    string
	Refresh   string
	Login     string
	Refreshed bool
}
