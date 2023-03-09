package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	fd "github.com/yitose/rssviewer/internal/feed"
)

type ItemTable struct {
	*tview.Table
}

func (i *ItemTable) GetItem(index int) (*fd.Item, error) {
	if i.GetRowCount() <= index {
		return nil, ErrFeedNotExist
	}
	res, ok := i.GetCell(index, 0).GetReference().(*fd.Item)
	if !ok {
		return nil, ErrFeedNotExist
	}
	return res, nil
}

func (t *ItemTable) setCell(i *fd.Item) {
	maxRow := t.GetRowCount()
	targetRow := maxRow
	for j := 0; j < maxRow; j++ {
		if t.GetCell(j, 0).Text == i.Title {
			return
		}
	}
	t.SetCell(targetRow, 0, tview.NewTableCell(i.Title).
		SetTextColor(tcell.Color(i.Color+1<<32)).
		SetReference(i))
}
