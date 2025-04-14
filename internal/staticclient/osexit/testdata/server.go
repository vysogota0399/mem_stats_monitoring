// Package testdata contains test data for the osexit analyzer.
package testdata

import (
	"os"
)

func Exiter() {
	os.Exit(-1)
}
