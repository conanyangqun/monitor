package highlight

// Segment represents a segment of text with highlight information
type Segment struct {
	Text        string `json:"text"`
	IsHighlight bool   `json:"is_highlight"`
}
