package boorufetch

import (
	"testing"
)

func TestFetch(t *testing.T) {
	posts, err := FromGelbooru("sakura_kyouko", 0)
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
		t.Logf("%v\n", p)
	}
}
