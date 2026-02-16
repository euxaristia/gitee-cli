package auth

import (
	"errors"

	keyring "github.com/zalando/go-keyring"
)

const (
	serviceName = "gitee-cli"
	userName    = "default"
)

func SaveToken(token string) error {
	return keyring.Set(serviceName, userName, token)
}

func LoadToken() (string, error) {
	token, err := keyring.Get(serviceName, userName)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return token, nil
}

func DeleteToken() error {
	err := keyring.Delete(serviceName, userName)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
