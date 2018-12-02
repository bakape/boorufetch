package boorufetch

type decoderCommon struct {
	Rating_          Rating `json:"rating"`
	File_url, Source string
}

func (d decoderCommon) Rating() Rating {
	return d.Rating_
}

func (d decoderCommon) FileURL() string {
	return d.File_url
}

func (d decoderCommon) SourceURL() string {
	return d.Source
}
