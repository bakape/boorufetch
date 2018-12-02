// Package boorufetch provides a unified interface for fetching posts from
// Gelbooru and Danbooru.

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

var (
	ratingRunes   = [...]byte{'s', 'q', 'e'}
	ratingStrings = [...]string{"safe", "questionable", "explicit"}
)

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

func (r Rating) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuoteRune(nil, rune(ratingRunes[int(r)])), nil
}

func (r Rating) String() string {
	return ratingStrings[int(r)]
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

// Single booru image post. Fields are lazily converted on demand for
// optimization purposes.
//
// Note that storing a large number of these in memory can be intensive both for
// memory consumption and garbage collection. In that case it is recomened to
// extract the required data from Post and allow the underlying structures to be
// deallocated.
type Post interface {
	// Return explicitness rating
	Rating() Rating

	// Return MD5 hash
	MD5() ([16]byte, error)

	// Return source file URL
	FileURL() string

	// Return sample file image URL or source file URL, if no sample present
	SampleURL() string

	// Return source URL, if any
	SourceURL() string

	// Return source file width
	Width() uint64

	// Return source file height
	Height() uint64

	// Return tags applied to post
	Tags() []Tag

	// Return last modification date
	UpdatedOn() (time.Time, error)

	// Return post creation date
	CreatedOn() (time.Time, error)
}
