package internal

import (
	"crypto/sha256"
	"encoding/hex"
)

type UrlHasherFunc func(input string) (string, error)

func Sha256UrlHasher(url string) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(url))
	if err != nil {
		return "", err
	}

	hashedBytes := hasher.Sum(nil)
	hashedString := hex.EncodeToString(hashedBytes)

	return hashedString, nil
}
