package boorufetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var gelbooruBalancer = newLoadBalancer("https", "gelbooru.com")

// Struct for decoding, augmenting and converting Gelbooru JSON responses
type gelbooruDecoder struct {
	decoderCommon
	tagParser
	Sample, isFetched                  bool
	Rating_                            Rating `json:"rating"`
	Change                             int64
	Height_                            uint64 `json:"height"`
	Width_                             uint64 `json:"width"`
	Hash, Directory, Created_at, Owner string
	updatedOnOverride                  time.Time
}

type tagParser interface {
	Tags() ([]Tag, error)
}

// Lazily fetch and parse the resource to speed up large page queries
func (d *gelbooruDecoder) fetch() (err error) {
	if d.isFetched {
		return nil
	}

	if d.Owner == "danbooru" {
		//  Fetch fresher tags and rating from Danbooru
		var danDec danbooruDecoder
		danDec, err = danbooruByMD5(d.Hash)
		if err != nil {
			return
		}
		d.Rating_, err = danDec.Rating()
		if err != nil {
			return
		}
		d.updatedOnOverride, err = danDec.UpdatedOn()
		if err != nil {
			return
		}
		d.tagParser = &danDec.danbooruTagDecoder
	} else {
		var r io.ReadCloser
		r, err = GelbooruFetchPage(fmt.Sprintf("md5:"+d.Hash), false, 0, 1)
		if err != nil {
			return
		}
		defer r.Close()
		d.tagParser, err = newGelbooruTagParser(r)
		if err != nil {
			return
		}
	}

	d.isFetched = true
	return
}

func (d *gelbooruDecoder) Tags() (t []Tag, err error) {
	err = d.fetch()
	if err != nil {
		return
	}
	return d.tagParser.Tags()
}

func (d *gelbooruDecoder) Rating() (r Rating, err error) {
	err = d.fetch()
	if err != nil {
		return
	}
	return d.Rating_, nil
}

func (d gelbooruDecoder) MD5() ([16]byte, error) {
	return decodeMD5(d.Hash)
}

func (d gelbooruDecoder) SampleURL() string {
	if d.Sample {
		return fmt.Sprintf(
			"https://simg3.gelbooru.com/samples/%s/sample_%s.jpg",
			d.Directory, d.Hash)
	}
	return d.File_url
}

func (d *gelbooruDecoder) UpdatedOn() (t time.Time, err error) {
	err = d.fetch()
	if err != nil {
		return
	}
	if d.updatedOnOverride.IsZero() {
		return time.Unix(d.Change, 0).UTC(), nil
	}
	return d.updatedOnOverride, nil
}

func (d gelbooruDecoder) CreatedOn() (time.Time, error) {
	return parseTime(time.RubyDate, d.Created_at)
}

func (d gelbooruDecoder) Width() uint64 {
	return d.Width_
}

func (d gelbooruDecoder) Height() uint64 {
	return d.Height_
}

// Makes request to a gelbooru page with selected tag query
func GelbooruFetchPage(query string, json bool, page, limit uint) (
	body io.ReadCloser, err error,
) {
	q := url.Values{
		"tags": {query},
		"pid":  {strconv.FormatUint(uint64(page), 10)},
	}
	if json {
		q.Set("page", "dapi")
		q.Set("s", "post")
		q.Set("q", "index")
		q.Set("limit", strconv.FormatUint(uint64(limit), 10))
		q.Set("json", "1")
	} else {
		q.Set("page", "post")
		q.Set("s", "list")
	}
	u := url.URL{
		Scheme:   "https",
		Host:     "gelbooru.com",
		Path:     "/index.php",
		RawQuery: q.Encode(),
	}
	return gelbooruBalancer.Fetch(u.String())
}

// Fetch posts from Gelbooru for the given tag query.
// If the source of the image is Danbooru, a more up-to-date set of tags is
// fetched from there.
// Fetches are limited to a maximum of FetcherCount concurrent requests to
// prevent antispam measures by the boorus.
func FromGelbooru(query string, page, limit uint) (posts []Post, err error) {
	r, err := GelbooruFetchPage(query, true, page, limit)
	if err != nil {
		return
	}
	defer r.Close()

	var dec []gelbooruDecoder
	err = json.NewDecoder(r).Decode(&dec)
	if err != nil || len(dec) == 0 {
		return
	}

	posts = make([]Post, len(dec))
	for i, d := range dec {
		posts[i] = &d
	}

	return
}

type gelbooruTagParser struct {
	root   *html.Node
	cached []Tag
}

func newGelbooruTagParser(r io.ReadCloser) (p *gelbooruTagParser, err error) {
	p = new(gelbooruTagParser)
	p.root, err = html.Parse(r)
	return
}

func (*gelbooruTagParser) getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// Find the tag list in the document using depth-first recursion
func (p *gelbooruTagParser) findTags(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "div" &&
		p.getAttr(n, "id") == "searchTags" {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		found := p.findTags(c)
		if found != nil {
			return found
		}
	}

	return nil
}

func (p *gelbooruTagParser) Tags() (tags []Tag, err error) {
	if p.cached != nil {
		return p.cached, nil
	}
	if p.root == nil {
		return
	}
	p.root = p.findTags(p.root)
	if p.root == nil {
		return
	}

	tags = make([]Tag, 0, 64)
	for n := p.root.FirstChild; n != nil; n = n.NextSibling {
		if n.Type != html.ElementNode || n.Data != "li" {
			continue
		}
		class := p.getAttr(n, "class")
		if !strings.HasPrefix(class, "tag-type-") {
			continue
		}

		// Get last <a> child
		var lastA *html.Node
		for n := n.LastChild; n != nil; n = n.PrevSibling {
			if n.Type == html.ElementNode && n.Data == "a" {
				lastA = n
				break
			}
		}
		if lastA == nil {
			continue
		}

		text := lastA.FirstChild
		if text == nil || text.Type != html.TextNode || text.Data == "" {
			continue
		}
		var tag Tag
		switch class {
		case "tag-type-artist":
			tag.Type = Author
		case "tag-type-character":
			tag.Type = Character
		case "tag-type-copyright":
			tag.Type = Series
		case "tag-type-metadata":
			tag.Type = Meta
		default:
			tag.Type = Undefined
		}

		tag.Tag = strings.Replace(html.UnescapeString(text.Data), " ", "_",
			-1)
		tags = append(tags, tag)
	}

	p.cached = tags
	p.root = nil
	return
}
