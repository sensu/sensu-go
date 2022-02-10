package secrets

import (
	"errors"
	"fmt"
)

var (
	ErrNoProviderDefined   = errors.New("no secret provider defined")
	ErrSecretsNotSupported = errors.New("secrets not supported")
)

type ErrInvalidSecretInfo string
type ErrProviderNotAvailable string
type ErrProviderNotFound string
type ErrSecretNotFound string

func (e ErrInvalidSecretInfo) Error() string {
	return fmt.Sprintf("invalid secret info: %s", string(e))
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
