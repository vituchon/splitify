// This package contains common code applicable to any type of interface
// (CLI, native GUI, web GUI, etc.), as well as foundational model code
// and utilities for interacting with external systems or dependencies (e.g., databases,
// network services, APIs, message brokers, filesystems, or cloud services).

package util

import (
	"math/rand"
	"os"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GenerateRandomNumber(min int, max int) int {
	return rand.Intn(max-min) + min
}
