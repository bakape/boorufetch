package boorufetch

import (
	"fmt"
)

type TagType uint8

const (
	Undefined TagType = iota
	Author
	Character
	Series
	Meta
)

// Explicitness rating of post
type Rating uint8

const (
	Safe Rating = iota
	Questionable
	Explicit
)

type ErrUnknownRating []byte

func (e ErrUnknownRating) Error() string {
	return fmt.Sprintf("unknown rating: `%s`", string(e))
}

func (r *Rating) UnmarshalJSON(buf []byte) (err error) {
	if len(buf) < 3 {
		return ErrUnknownRating(buf)
	}
	switch buf[1] {
	case 's':
		*r = Safe
	case 'q':
		*r = Questionable
	case 'e':
		*r = Explicit
	default:
		return ErrUnknownRating(buf)
	}
	return nil
}

// Number of parallel fetch workers
const FetcherCount = 4

type Tag struct {
	Type TagType `json:"type"`
	Tag  string  `json:"tag"`
}

type Post struct {
	Rating    Rating `json:"rating"`
	Sample    bool   `json:"sample"`
	ChangedOn int64  `json:"change"`
	Hash      string `json:"hash"`
	Owner     string `json:"owner"`
	FileURL   string `json:"file_url"`
	Directory string `json:"directory"`
	Tags      []Tag  `json:"-"`
}
