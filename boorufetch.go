package boorufetch

import (
	"fmt"
	"strconv"
	"time"
)

// Number of parallel fetch workers
const FetcherCount = 4

// Tag category
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

var ratingRunes = [...]byte{'s', 'q', 'e'}

func (r *Rating) UnmarshalJSON(buf []byte) error {
	if len(buf) < 3 {
		return ErrUnknownRating(buf)
	}
	for i := 0; i < 3; i++ {
		if buf[1] == ratingRunes[i] {
			*r = Rating(i)
			return nil
		}
	}
	return ErrUnknownRating(buf)
}

func (r *Rating) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuoteRune(nil, rune(ratingRunes[int(*r)])), nil
}

type ErrUnknownRating []byte

func (e ErrUnknownRating) Error() string {
	return fmt.Sprintf("unknown rating: `%s`", string(e))
}

// Tag associated to a post
type Tag struct {
	Type TagType `json:"type"`
	Tag  string  `json:"tag"`
}

// Single booru image post
type Post struct {
	Rating    Rating    `json:"rating"`
	MD5       [16]byte  `json:"md5"`
	FileURL   string    `json:"file_url"`
	SampleURL string    `json:"sample_url"`
	Source    string    `json:"source"`
	Tags      []Tag     `json:"tags"`
	UpdatedOn time.Time `json:"updated_on"`
	CreatedOn time.Time `json:"created_on"`
}
