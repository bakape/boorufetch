package boorufetch

import "testing"

func TestDanbooruFetch(t *testing.T) {
	posts, err := FromDanbooru("sakura_kyouko", 0, 5)
	logPosts(t, posts, err)
}
