package main

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// based on google/uuid
func getUUIDString() string {
	uuid := uuid.New()
	return fmt.Sprintf("%016x", uuid[0:16])
}

// newUUID generates a random UUID according to RFC 4122
func newUUIDString() string {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return ""
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	// return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
	return fmt.Sprintf("%016x", uuid)
}
