package boorufetch

var danbooruBalancer = newLoadBalancer("https", "danbooru.donmai.us")

func danbooruFetchTags(md5 string) (tags []Tag, rating Rating, err error) {
	panic("TODO")
}
