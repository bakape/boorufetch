package boorufetch

import (
	"testing"
)

func TestGelbooruFetch(t *testing.T) {
	posts, err := FromGelbooru("sakura_kyouko", 0, 3)
	logPosts(t, posts, err)
}

func logPosts(t *testing.T, posts []Post, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
	if len(posts) == 0 {
		t.Fatal("no images found")
	}
	for i, p := range posts {
		if i == 10 {
			break
		}
		t.Logf("\t\n%v\n%v\n", p, p.Tags())
	}
}

func TestDanbooruFetch(t *testing.T) {
	posts, err := FromDanbooru("sakura_kyouko", 0, 3)
	logPosts(t, posts, err)
}
