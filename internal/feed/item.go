package feed

import (
	"github.com/mmcdole/gofeed"
)

type Item struct {
	*gofeed.Item
	Belong string
	Color  int
}
