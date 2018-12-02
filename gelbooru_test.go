package boorufetch

import "testing"

func TestGelbooruFetch(t *testing.T) {
	posts, err := FromGelbooru("sakura_kyouko", 0, 5)
	logPosts(t, posts, err)
}
