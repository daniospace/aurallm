package proxy

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashKey(key string) string {
	h := sha256.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}
