package crypto

import "crypto/sha1"

// GenerateSHA1ID returns the SHA1 hash of the input string.
// Used to generate a unique identifier nodes in the DHT.
func GenerateSHA1ID(val string) string {
	hash := sha1.New()
	hash.Write([]byte(val))
	return string(hash.Sum(nil))
}
