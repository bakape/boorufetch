package boorufetch

import "testing"

func TestGelbooruFetch(t *testing.T) {
	posts, err := FromGelbooru("sakura_kyouko", 0, 5)
	logPosts(t, err, posts...)
}

func TestGelbooruNoMatch(t *testing.T) {
	posts, err := FromGelbooru("md5:571fbb01d157f7eb62d8f4b3b7250ad4", 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 0 {
		t.Fatal("expected no posts")
	}
}
