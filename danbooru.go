package boorufetch

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	danbooruBalancer = newLoadBalancer("https", "danbooru.donmai.us")
	logRequests      = os.Getenv("LOG_REQUESTS") == "1"
)

type danbooruDecoder struct {
	decoderCommon
	danbooruTagDecoder
	Has_large                              bool
	Rating_                                Rating `json:"rating"`
	Image_width, Image_height              uint64
	Large_file_url, Updated_at, Created_at string
	Hash                                   string `json:"md5"`
}

func (d danbooruDecoder) Rating() (Rating, error) {
	return d.Rating_, nil
}

func (d danbooruDecoder) MD5() ([16]byte, error) {
	return DecodeMD5(d.Hash)
}

func (d danbooruDecoder) SampleURL() string {
	if d.Has_large {
		return d.Large_file_url
	}
	return d.File_url
}

func (d danbooruDecoder) UpdatedOn() (time.Time, error) {
	return parseTime(time.RFC3339Nano, d.Updated_at)
}

func (d danbooruDecoder) CreatedOn() (time.Time, error) {
	return parseTime(time.RFC3339Nano, d.Created_at)
}

func (d danbooruDecoder) Width() uint64 {
	return d.Image_width
}

func (d danbooruDecoder) Height() uint64 {
	return d.Image_height
}

type danbooruTagDecoder struct {
	Tag_string_general, Tag_string_character, Tag_string_copyright string
	Tag_string_artist, Tag_string_meta                             string
	cached                                                         []Tag
}

func (d *danbooruTagDecoder) parse(typ TagType, s *string) {
	for _, t := range strings.Split(*s, " ") {
		if len(t) == 0 {
			continue
		}
		d.cached = append(d.cached, Tag{
			Type: typ,
			Tag:  t,
		})
	}

	*s = "" // Free up memory
}

func (d *danbooruTagDecoder) Tags() ([]Tag, error) {
	if d.cached != nil {
		return d.cached, nil
	}
	d.cached = make([]Tag, 0, 64)

	d.parse(Author, &d.Tag_string_artist)
	d.parse(Character, &d.Tag_string_character)
	d.parse(Series, &d.Tag_string_copyright)
	d.parse(Undefined, &d.Tag_string_general)
	d.parse(Meta, &d.Tag_string_meta)

	dedupTags(&d.cached)
	return d.cached, nil
}

func danbooruURL(q url.Values) string {
	u := url.URL{
		Scheme:   "https",
		Host:     "danbooru.donmai.us",
		Path:     "/posts.json",
		RawQuery: q.Encode(),
	}
	return u.String()
}

// Fetch posts from Danbooru for the given tag query.
// Note the query may only contain up to 2 tags.
// Fetches are limited to a maximum of FetcherCount concurrent requests to
// prevent antispam measures by the boorus.
func FromDanbooru(query string, page, limit uint) (posts []Post, err error) {
	u := danbooruURL(url.Values{
		"tags":  {query},
		"page":  {strconv.FormatUint(uint64(page), 10)},
		"limit": {strconv.FormatUint(uint64(limit), 10)},
	})
	r, err := danbooruBalancer.Fetch(u)
	if err != nil {
		return
	}
	defer r.Close()

	if logRequests {
		var buf []byte
		buf, err = ioutil.ReadAll(r)
		if err != nil {
			return
		}
		fmt.Printf("fetched from %s: %s\n", u, string(buf))

		r.Close()
		r = dummyCLoser{bytes.NewReader(buf)}
	}
	var dec []danbooruDecoder
	err = json.NewDecoder(r).Decode(&dec)
	switch err {
	case nil:
	case io.EOF:
		err = nil
		return
	default:
		return
	}
	if len(dec) == 0 {
		return
	}

	posts = make([]Post, len(dec))
	for i := range dec {
		posts[i] = &dec[i]
	}
	return
}

// Fetch a single danbooru post by MD5
func danbooruByMD5(md5 string) (dec danbooruDecoder, err error) {
	r, err := danbooruBalancer.Fetch(danbooruURL(url.Values{
		"md5": {md5},
	}))
	if err != nil {
		return
	}
	defer r.Close()
	err = json.NewDecoder(r).Decode(&dec)
	return
}

// Fetch a single danbooru post by MD5. If no match found, Post will be nil.
func DanbooruByMD5(md5 [16]byte) (Post, error) {
	d, err := danbooruByMD5(hex.EncodeToString(md5[:]))
	if d.FileURL() == "" { // Zero value check
		return nil, err
	}
	return &d, err
}
