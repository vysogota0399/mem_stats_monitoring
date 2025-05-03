package crypto

import (
	"crypto/subtle"
	"fmt"
	"hash"
	"io"
)

// Cms represents a cryptographic message signing system that uses a specified hash algorithm
// for creating and verifying message signatures. It implements a simple CMS-style signing mechanism.
type Cms struct {
	signatureAlg hash.Hash
}

// NewCms creates a new Cms instance with the specified hash algorithm.
// The hash algorithm will be used for both signing and verification operations.
//
// Example:
//
//	cms := NewCms(sha256.New())
func NewCms(signAlg hash.Hash) *Cms {
	return &Cms{
		signatureAlg: signAlg,
	}
}

// Sign generates a cryptographic signature for the provided message.
// The message is read from the provided io.Reader and hashed using the configured hash algorithm.
//
// Returns:
//   - []byte: The generated signature
//   - error: Any error that occurred during the signing process
//
// Example:
//
//	signature, err := cms.Sign(strings.NewReader("message to sign"))
func (c Cms) Sign(msg io.Reader) ([]byte, error) {
	h := c.signatureAlg
	if _, err := io.Copy(h, msg); err != nil {
		return nil, fmt.Errorf("internal/utils/crypto/cms.go add data to hash failed %w", err)
	}

	return h.Sum(nil), nil
}

// Verify checks if the provided message matches the expected signature.
// The message is read from the provided io.Reader and hashed using the configured hash algorithm.
// The resulting hash is compared with the expected signature using constant-time comparison
// to prevent timing attacks.
//
// Returns:
//   - bool: true if the signature is valid, false otherwise
//   - error: Any error that occurred during the verification process
//
// Example:
//
//	valid, err := cms.Verify(strings.NewReader("message to verify"), signature)
func (c Cms) Verify(data io.Reader, expectedSign []byte) (bool, error) {
	sign, err := c.Sign(data)
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(sign, expectedSign) == 1, nil
}
