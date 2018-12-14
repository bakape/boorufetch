package boorufetch

import (
	"encoding/hex"
	"fmt"
	"time"
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

func parseTime(layout, s string) (t time.Time, err error) {
	if s == "" {
		return
	}
	t, err = time.Parse(layout, s)
	t = t.Round(time.Second).UTC()
	return
}

// Deduplicate tags array
func dedupTags(tags *[]Tag) {
	tmp := make(map[Tag]struct{}, len(*tags))

	for _, t := range *tags {
		tmp[t] = struct{}{}
	}
	*tags = (*tags)[:0]
	for t := range tmp {
		*tags = append(*tags, t)
	}
}
