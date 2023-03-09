package tui

import (
	"github.com/pkg/errors"
	"github.com/rivo/tview"

	fd "github.com/yitose/rssviewer/internal/feed"
)

var ErrGroupNotExist = errors.Errorf("Feed Not Exist")

type GroupTable struct {
	*FeedTable
}

func (t *GroupTable) setCell(g *fd.Group) *tview.TableCell {
	if g == nil {
		return nil
	}

	maxRow := t.GetRowCount()
	targetRow := maxRow
	for i := 0; i < maxRow; i++ {
		cell := t.GetCell(i, 0)
		ref, ok := cell.GetReference().(*GroupCellRef)
		if ok {
			if ref.Group.Title == g.Title {
				targetRow = i
				break
			}
		}
	}

	cell := tview.NewTableCell(g.Title).SetReference(NewGroupCellRef(g))

	t.SetCell(targetRow, 0, cell)

	// SelectionChangedFuncを発火する
	if maxRow == 0 {
		t.Select(t.GetSelection())
	}

	return cell
}
