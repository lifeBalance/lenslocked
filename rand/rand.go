package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

/*
RandomBytes returns a slice of `sliceSize` cryptographically secure random bytes.

The function ensures exactly `sliceSize` bytes are read; if fewer are available or an error occurs during reading, it returns an error wrapped with context.
Parameters:
  - sliceSize: The number of random bytes to generate (must be non-negative).
Returns:
  - []byte: A slice of exactly `sliceSize` random bytes on success.
  - error: Wrapped error if reading fails or insufficient bytes are read.
Example:

  key, err := Bytes(32)  // Generate a 256-bit (32-byte) random key
  if err != nil {
      log.Fatal(err)
  }
  fmt.Printf("Random key: %x\n", key)  // "a1b2c3d4e5f6..." (32 hex bytes)
*/

func RandomBytes(sliceSize int) ([]byte, error) {
	b := make([]byte, sliceSize)
	numBytesRead, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("bytes: %w", err)
	}
	if numBytesRead < sliceSize {
		return nil, fmt.Errorf("bytes: didn't read enough random bytes")
	}
	return b, nil
}

/*
RandomBase64String returns a URL-safe base64-encoded string generated
from `len` cryptographically secure random bytes.

Notes:
  - The `len` parameter is the number of raw random bytes; the resulting
    base64 string will be longer (roughly 4/3 of `len`) and may include
    padding characters ('=') because URLEncoding is used.
  - Uses crypto/rand under the hood via RandomBytes, suitable for tokens,
    secrets, and CSRF keys.
  - Returns a wrapped error if random byte generation fails.

Example:

	s, err := RandomBase64String(32) // 32 bytes -> ~43 chars (with padding)
	if err != nil { log.Fatal(err) }
	fmt.Println(s)
*/
func RandomBase64String(len int) (string, error) {
	randomBytes, err := RandomBytes(len)
	if err != nil {
		log.Fatal(err)
		return "", fmt.Errorf("RandomBase64String: %w", err)
	}
	return base64.URLEncoding.EncodeToString(randomBytes), nil
}
