package crypto

import (
	"crypto/subtle"
	"fmt"
	"hash"
	"io"
)

type Cms struct {
	signatureAlg hash.Hash
}

func NewCms(signAlg hash.Hash, key []byte) *Cms {
	return &Cms{
		signatureAlg: signAlg,
	}
}

func (c Cms) Sign(msg io.Reader) ([]byte, error) {
	if _, err := io.Copy(c.signatureAlg, msg); err != nil {
		return nil, fmt.Errorf("internal/utils/crypto/cms.go add data to hash failed %w", err)
	}

	return c.signatureAlg.Sum(nil), nil
}

func (c Cms) Verify(data io.Reader, expectedSign []byte) (bool, error) {
	sign, err := c.Sign(data)
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(sign, expectedSign) == 1, nil
}
