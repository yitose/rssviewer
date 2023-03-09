package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/pkg/errors"
	"github.com/rivo/tview"
	fd "github.com/yitose/rssviewer/internal/feed"
)

type FeedTable struct {
	*tview.Table
}

var ErrFeedNotExist = errors.Errorf("Feed Not Exist")

func (t *FeedTable) setCell(f *fd.Feed) *tview.TableCell {
	if f == nil {
		return nil
	}

	maxRow := t.GetRowCount()
	targetRow := maxRow
	for i := 0; i < maxRow; i++ {
		cell := t.GetCell(i, 0)
		ref, ok := cell.GetReference().(*FeedCellRef)
		if ok {
			if ref.Feed.FeedLink == f.FeedLink {
				targetRow = i
				break
			}
		}
	}

	cell := tview.NewTableCell(f.Title).
		SetTextColor(tcell.Color(f.Color + 1<<32)).
		SetReference(NewFeedCellRef(f))

	t.SetCell(targetRow, 0, cell)

	// SelectionChangedFuncを発火する
	if maxRow == 0 {
		t.Select(t.GetSelection())
	}

	return cell
}
