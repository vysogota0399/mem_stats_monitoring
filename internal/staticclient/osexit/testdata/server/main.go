// Package main contains test data for the osexit analyzer.
package main

import (
	"os"
)

func exiter() {
	os.Exit(-1) // want `os.Exit call found - this may cause unexpected program termination`
}

func main() {
	exiter()
}
