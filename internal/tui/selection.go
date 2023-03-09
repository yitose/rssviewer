package tui

import (
	"fmt"
	"time"

	db "github.com/yitose/rssviewer/internal/db"
	fd "github.com/yitose/rssviewer/internal/feed"
)

func (t *Tui) commonKeyHelp() [][]string {
	help := [][]string{
		{"n", "new"},
		{"i", "import"},
	}

	if t.FeedWidget.GetRowCount() > 0 {
		help = append(help, [][]string{
			{"e", "export"},
			{"R", "update"},
			{"D", "description"},
		}...)
	}

	help = append(help, []string{"q", "quit"})

	return help
}

func (t *Tui) setSelectionFunc() {
	t.GroupWidget.Table.SetSelectionChangedFunc(t.groupTableSelectionChangedFunc)
	t.FeedWidget.Table.SetSelectionChangedFunc(t.feedTableSelectionChangedFunc)
	t.ItemWidget.SetSelectionChangedFunc(t.itemTableSelectionChangedFunc)
}

func (t *Tui) groupTableSelectionChangedFunc(row, column int) {
	if !t.GroupWidget.HasFocus() {
		return
	}

	help := [][]string{
		{"l", "→"},
		{"J", "↓"},
	}

	if rowCount := t.GroupWidget.GetRowCount(); rowCount == 0 || row >= rowCount {
		t.Help(append(help, t.commonKeyHelp()...))
		return
	}

	help = append(help, []string{"d", "delete"})
	help = append(help, []string{"\n", ""})
	t.Help(append(help, t.commonKeyHelp()...))

	cellRef := t.GroupWidget.GetCell(row, column).GetReference().(*GroupCellRef)
	group := cellRef.Group

	desc := [][]string{
		{"Title", group.Title},
	}
	t.Descript(desc)

	items := []*fd.Item{}

	if group.Title == db.TodaysFeedTitle {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		for i := 0; i < t.FeedWidget.GetRowCount(); i++ {
			ref, ok := t.FeedWidget.GetCell(i, 0).GetReference().(*FeedCellRef)
			if ok {
				for _, i := range ref.Feed.Items {
					if i.PublishedParsed != nil && !today.After(*i.PublishedParsed) {
						items = append(items, i)
					}
				}
			}
		}
	} else {
		for _, link := range group.FeedLinks {
			for _, f := range t.DB.Feed {
				if f.FeedLink == link {
					items = append(items, f.Items...)
				}
			}
		}
	}

	t.ItemWidget.Clear()

	fd.SortItems(items)

	for _, i := range items {
		t.ItemWidget.setCell(i)
	}

	t.ItemWidget.ScrollToBeginning().Select(cellRef.Cursor, 0)

	t.ConfirmationStatus = defaultConfirmationStatus
}

func (t *Tui) feedTableSelectionChangedFunc(row, column int) {
	if !t.FeedWidget.HasFocus() {
		return
	}

	help := [][]string{
		{"l", "→"},
		{"K", "↑"},
	}

	if rowCount := t.FeedWidget.GetRowCount(); rowCount == 0 || row >= rowCount {
		help = append(help, []string{"\n", ""})
		t.Help(append(help, t.commonKeyHelp()...))
		return
	}

	help = append(help, [][]string{
		{"c", "recolor"},
		{"d", "delete"},
		{"v", "select"},
		{"m", "make"},
	}...)
	help = append(help, []string{"\n", ""})
	t.Help(append(help, t.commonKeyHelp()...))

	t.ItemWidget.Clear()

	cellRef := t.FeedWidget.GetCell(row, column).GetReference().(*FeedCellRef)
	feed := cellRef.Feed

	t.ItemWidget.ScrollToBeginning().Select(cellRef.Cursor, 0)

	desc := [][]string{
		{"Title", feed.Title},
		{"Description", feed.Description},
		{"PubLished", fd.FormatDate(feed.PublishedParsed)},
		{"ColorCode", fmt.Sprint(feed.Color)},
		{"URL", feed.FeedLink},
	}
	t.Descript(desc)

	for _, i := range feed.Items {
		t.ItemWidget.setCell(i)
	}

	t.ConfirmationStatus = defaultConfirmationStatus
}

func (t *Tui) itemTableSelectionChangedFunc(row, column int) {
	if !t.ItemWidget.HasFocus() {
		return
	}

	help := [][]string{
		{"h", "←"},
		{"l", "↓"},
	}

	if rowCount := t.ItemWidget.GetRowCount(); rowCount == 0 || row >= rowCount {
		help = append(help, []string{"\n", ""})
		t.Help(append(help, t.commonKeyHelp()...))
		return
	}

	help = append(help, [][]string{
		{"o", "open"},
		{"c", "recolor"},
	}...)
	help = append(help, []string{"\n", ""})
	t.Help(append(help, t.commonKeyHelp()...))

	switch t.CurrentLeftTable {
	case enumGroupWidget:
		cell := t.GroupWidget.GetCell(t.GroupWidget.GetSelection())
		cellRef, ok := cell.GetReference().(*GroupCellRef)
		if ok {
			cellRef.Cursor = row
		}
	case enumFeedWidget:
		cell := t.FeedWidget.GetCell(t.FeedWidget.GetSelection())
		cellRef, ok := cell.GetReference().(*FeedCellRef)
		if ok {
			cellRef.Cursor = row
		}
	default:
		return
	}

	cell := t.ItemWidget.GetCell(row, column)
	item := cell.GetReference().(*fd.Item)

	author := ""
	if item.Author != nil {
		author = item.Author.Name
	}

	desc := [][]string{}
	if t.CurrentLeftTable == enumGroupWidget {
		desc = append(desc, []string{"Feed", t.DB.GetItemParent(item).Title})
	}
	desc = append(desc, [][]string{
		{"Title", item.Title},
		{"Description", item.Description},
		{"PubLished", fd.FormatDate(item.PublishedParsed)},
		{"Author", author},
		{"Link", item.Link},
	}...)

	t.Descript(desc)

	t.ConfirmationStatus = defaultConfirmationStatus
}
