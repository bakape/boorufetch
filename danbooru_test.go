package boorufetch

import "testing"

func TestDanbooruFetch(t *testing.T) {
	posts, err := FromDanbooru("sakura_kyouko", 0, 5)
	logPosts(t, err, posts...)
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

func TestDanbooruByMD5(t *testing.T) {
	hash, err := DecodeMD5("39b1f4f5298c446b483b858335d85fc7")
	if err != nil {
		t.Fatal(err)
	}
	post, err := DanbooruByMD5(hash)
	logPosts(t, err, post)
}

func TestDanbooruByMD5NoPost(t *testing.T) {
	hash, err := DecodeMD5("39b1f4f5298c446b483c858335d85fc7")
	if err != nil {
		t.Fatal(err)
	}
	post, err := DanbooruByMD5(hash)
	if err != nil {
		t.Fatal(err)
	}
	if post != nil {
		t.Fatal("expected no posts")
	}
}
