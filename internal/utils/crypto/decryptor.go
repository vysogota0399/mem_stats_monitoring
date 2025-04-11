package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
)

type Decryptor struct {
	privateKey io.Reader
}

func NewDecryptor(pk io.Reader) *Decryptor {
	return &Decryptor{privateKey: pk}
}

func (d *Decryptor) Decrypt(ciphertext string) (string, error) {
	pkdata, err := io.ReadAll(d.privateKey)
	if err != nil {
		return "", fmt.Errorf("decryptor: failed to read private key: %w", err)
	}

	block, _ := pem.Decode(pkdata)
	if block == nil {
		return "", fmt.Errorf("decryptor: failed to decode private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("decryptor: failed to parse private key: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("decryptor: private key is not a RSA key")
	}

	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decryptor: failed to decode ciphertext: %w", err)
	}

	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, decoded, nil)
	if err != nil {
		return "", fmt.Errorf("decryptor: failed to decrypt ciphertext: %w", err)
	}

	return string(decrypted), nil
}
