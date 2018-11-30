package boorufetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var gelbooruBalancer = newLoadBalancer("https", "gelbooru.com")

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
	defer func() {
		if r != nil {
			r.Close()
		}
	}()
	err = json.NewDecoder(r).Decode(&posts)
	if err != nil || len(posts) == 0 {
		return
	}

	for i := range posts {
		// TODO: Fetch fresher tags and rating from Danbooru

		if r != nil {
			r.Close()
			r = nil
		}
		r, err = GelbooruFetchPage(fmt.Sprintf("md5:"+posts[i].Hash), false, 0)
		if err != nil {
			return
		}
		posts[i].Tags, err = gelbooruParseTags(r)
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
		tag.Tag = html.UnescapeString(text.Data)
		*t = append(*t, tag)
	}
}
