package search

type SearchMode int

const (
	Filename SearchMode = iota
	Content
)

func (m SearchMode) String() string {
	switch m {
	case Filename:
		return "filename"
	case Content:
		return "content"
	}
	panic("unhandled SearchMode")
}

func (m SearchMode) Toggle() SearchMode {
	if m == Filename {
		return Content
	}

	return Filename
}
