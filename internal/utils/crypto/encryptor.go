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

type Encryptor struct {
	certData io.Reader
}

func NewEncryptor(cert io.Reader) *Encryptor {
	return &Encryptor{certData: cert}
}

func (e *Encryptor) Encrypt(message []byte) (string, error) {
	certBytes, err := io.ReadAll(e.certData)
	if err != nil {
		return "", fmt.Errorf("encryptor: failed to read public key: %w", err)
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return "", fmt.Errorf("encryptor: failed to decode public key")
	}
	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("encryptor: failed to parse public key: %w", err)
	}

	publicKey, ok := parsedCert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("encryptor: failed to cast public key to rsa.PublicKey")
	}

	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, message, nil)
	if err != nil {
		return "", fmt.Errorf("encryptor: failed to encrypt message: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}
