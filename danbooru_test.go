package boorufetch

import "testing"

func TestDanbooruFetch(t *testing.T) {
	posts, err := FromDanbooru("sakura_kyouko", 0, 5)
	logPosts(t, posts, err)
}

func TestDanbooruNoMatch(t *testing.T) {
	posts, err := FromDanbooru("md5:571fbb01d157f7eb62d8f4b3b7250ad4", 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 0 {
		t.Fatal("expected no posts")
	}
}
