package boorufetch

type decoderCommon struct {
	File_url, Source string
}

func (d decoderCommon) FileURL() string {
	return d.File_url
}

func (d decoderCommon) SourceURL() string {
	return d.Source
}
