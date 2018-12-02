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
type GelbooruDecoder struct {
	Rating    Rating `json:"rating"`
	Sample    bool   `json:"sample"`
	ChangedOn int64  `json:"change"`
	MD5       string `json:"hash"`
	FileURL   string `json:"file_url"`
	Directory string `json:"directory"`
	Tags      []Tag  `json:"-"`
	CreatedOn string `json:"created_at"`
	Source    string `json:"source"`
}

// Fetch more detailed tags than availbale from the JSON API
func (d *GelbooruDecoder) FetchTags() (err error) {
	r, err := GelbooruFetchPage(fmt.Sprintf("md5:"+d.MD5), false, 0)
	if err != nil {
		return
	}
	defer r.Close()
	d.Tags, err = gelbooruParseTags(r)
	return
}

// Convert to Post
func (d GelbooruDecoder) ToPost() (p Post, err error) {
	p.Rating = d.Rating
	p.Tags = d.Tags
	p.Source = d.Source
	p.MD5, err = decodeMD5(d.MD5)
	if err != nil {
		return
	}

	p.CreatedOn, err = time.Parse(time.RubyDate, d.CreatedOn)
	if err != nil {
		return
	}
	p.CreatedOn = p.CreatedOn.UTC()
	p.UpdatedOn = time.Unix(d.ChangedOn, 0).UTC()

	p.FileURL = d.FileURL
	if d.Sample {
		p.SampleURL = fmt.Sprintf(
			"https://simg3.gelbooru.com/samples/%s/sample_%s.jpg",
			d.Directory, d.MD5)
	} else {
		p.SampleURL = p.FileURL
	}

	return
}

// Makes request to a gelbooru page with selected tag query
func GelbooruFetchPage(query string, json bool, page uint) (
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
		q.Set("limit", "100")
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
func FromGelbooru(query string, page uint) (posts []Post, err error) {
	r, err := GelbooruFetchPage(query, true, page)
	if err != nil {
		return
	}
	defer r.Close()

	var dec []GelbooruDecoder
	err = json.NewDecoder(r).Decode(&dec)
	if err != nil || len(dec) == 0 {
		return
	}

	posts = make([]Post, len(dec))
	for i := range dec {
		// TODO: Fetch fresher tags and rating from Danbooru

		err = dec[i].FetchTags()
		if err != nil {
			return
		}
		posts[i], err = dec[i].ToPost()
		if err != nil {
			return
		}
	}

	return
}

func gelbooruParseTags(r io.ReadCloser) (tags []Tag, err error) {
	doc, err := html.Parse(r)
	if err != nil {
		return
	}
	tagList := gelbooruFindTags(doc)
	if tagList == nil {
		return
	}
	gelbooruScrapeTags(tagList, &tags)
	return
}

// Find the tag list in the document using depth-first recursion
func gelbooruFindTags(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "div" &&
		getAttr(n, "id") == "searchTags" {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		found := gelbooruFindTags(c)
		if found != nil {
			return found
		}
	}

	return nil
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func gelbooruScrapeTags(n *html.Node, t *[]Tag) {
	for n := n.FirstChild; n != nil; n = n.NextSibling {
		if n.Type != html.ElementNode || n.Data != "li" {
			continue
		}
		class := getAttr(n, "class")
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

		tag.Tag = strings.Replace(html.UnescapeString(text.Data), " ", "_", -1)
		*t = append(*t, tag)
	}
}
