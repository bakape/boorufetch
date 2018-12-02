package boorufetch

import (
	"encoding/hex"
	"fmt"
)

// Decode hex string to MD5 hash
func decodeMD5(s string) (buf [16]byte, err error) {
	n, err := hex.Decode(buf[:], []byte(s))
	if err != nil {
		return
	}
	if n != 16 {
		err = fmt.Errorf("invalid MD5 hash: `%s`", err)
	}
	return
}