// Package crypto provides cryptographic utilities for message signing and verification.
//
// The package implements a simple CMS mechanism
// that allows for signing and verifying messages using various hash algorithms.
//
// Example usage:
//
//	// Create a new CMS instance with SHA256
//	cms := crypto.NewCms(sha256.New())
//
//	// Sign a message
//	signature, err := cms.Sign(strings.NewReader("message to sign"))
//	if err != nil {
//		// handle error
//	}
//
//	// Verify a message
//	valid, err := cms.Verify(strings.NewReader("message to verify"), signature)
//	if err != nil {
//		// handle error
//	}
//	if valid {
//		// signature is valid
//	}
package crypto
