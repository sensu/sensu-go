package secrets

import "fmt"

type ErrInvalidSecretInfo string
type ErrNoProviderDefined struct{}
type ErrProviderNotAvailable string
type ErrProviderNotFound string
type ErrSecretNotFound string
type ErrSecretsNotSupported struct{}

func (e ErrInvalidSecretInfo) Error() string {
	return fmt.Sprintf("invalid secret info: %s", string(e))
}

func (e ErrNoProviderDefined) Error() string {
	return "no secret provider defined"
}

func (e ErrProviderNotAvailable) Error() string {
	return fmt.Sprintf("secret provider not available: %s", string(e))
}

func (e ErrProviderNotFound) Error() string {
	return fmt.Sprintf("secret provider not found: %s", string(e))
}

func (e ErrSecretNotFound) Error() string {
	return fmt.Sprintf("secret not found: %s", string(e))
}

func (e ErrSecretsNotSupported) Error() string {
	return "secrets not supported"
}
