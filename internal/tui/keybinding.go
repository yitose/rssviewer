package tui

import (
	"errors"
	"os"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	db "github.com/yitose/rssviewer/internal/db"
	fd "github.com/yitose/rssviewer/internal/feed"
)

const msgRefusedByLoading = "It is not allowed during loading."

func (t *Tui) setKeyBinding() {
	t.App.SetInputCapture(t.appInputCaptureFunc)
	t.GroupWidget.SetInputCapture(t.groupTableInputCaptureFunc)
	t.FeedWidget.SetInputCapture(t.feedTableInputCaptureFunc)
	t.ItemWidget.SetInputCapture(t.itemTableInputCaptureFunc)
	t.InputWidget.SetInputCapture(t.inputWidgetInputCaptureFunc)
	t.DescriptionWidget.SetInputCapture(t.descriptionWidgetInputCaptureFunc)
	t.ColorWidget.SetInputCapture(t.colorWidgetInputCaptureFunc)
}

func (t *Tui) appInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	if t.InputWidget.HasFocus() {
		return event
	}

	switch event.Key() {

	}

	switch event.Rune() {
	case 'e':
		if t.IsLoading {
			t.Notify(msgRefusedByLoading, true)
		} else {
			if t.ConfirmationStatus == 'e' {
				listFile, err := os.Create(db.ExportListPath)
				if err != nil {
					panic(err)
				}
				defer listFile.Close()
				for _, f := range t.DB.Feed {
					if _, err := listFile.WriteString(f.FeedLink + "\n"); err != nil {
						panic(err)
					}
				}
				t.Notify("Exported to "+db.ExportListPath+".", false)
				t.ConfirmationStatus = defaultConfirmationStatus
			} else {
				t.Notify("Press e again to export feed urls.", false)
				t.ConfirmationStatus = 'e'
			}
		}
	case 'q':
		t.App.Stop()
	case 'n':
		if t.IsLoading {
			t.Notify(msgRefusedByLoading, true)
		} else {
			t.InputWidget.SetTitle("New Feed")
			t.InputWidget.Mode = 'n'
			t.Pages.ShowPage(inputField)
			t.App.SetFocus(t.InputWidget)
			t.Notify("Enter a feed URL or a command to output feed as xml.", false)
			return nil
		}
	case 'i':
		if t.IsLoading {
			t.Notify(msgRefusedByLoading, true)
		} else {
			if t.ConfirmationStatus == 'i' {
				go func() {
					if err := t.AddFeedsFromURL(db.ImportListPath); err != nil {
						if err == ErrImportFileNotFound {
							t.Notify("import failed:"+err.Error(), true)
						} else {
							panic(err)
						}
					} else {
						t.Notify("Imported from "+db.ImportListPath+".", false)
					}
					t.App.QueueUpdateDraw(func() {})
				}()
				t.ConfirmationStatus = defaultConfirmationStatus
			} else {
				t.Notify("Press i again to import from "+db.ImportListPath+".", false)
				t.ConfirmationStatus = 'i'
			}
		}
	case 'D':
		if t.IsLoading {
			t.Notify(msgRefusedByLoading, true)
		} else {
			t.Pages.ShowPage(descriptionField)
			t.App.SetFocus(t.DescriptionWidget)
		}
	case 'R':
		if t.IsLoading {
			t.Notify(msgRefusedByLoading, true)
		} else {
			go func() {
				if err := t.UpdateAllFeed(); err != nil {
					panic(err)
				}
				t.App.QueueUpdateDraw(func() {})
			}()
		}
	}

	return event
}

func (t *Tui) groupTableInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {

	}

	switch event.Rune() {
	case 'd':
		if t.IsLoading {
			t.Notify(msgRefusedByLoading, true)
		} else {
			if t.ConfirmationStatus == 'd' {
				cell := t.GroupWidget.GetCell(t.GroupWidget.GetSelection())
				ref, ok := cell.GetReference().(*GroupCellRef)
				if ok {
					if ref.Group.Title == db.TodaysFeedTitle {
						t.Notify(db.TodaysFeedTitle+" is an automatically generated group, and cannot be removed.", true)
					} else {
						if err := t.DB.DeleteGroup(ref.Group); err != nil {
							panic(err)
						}
						t.Notify("deleted.", false)
					}
				} else {
					t.Notify("delete failed.", true)
				}
				t.ConfirmationStatus = defaultConfirmationStatus
				t.resetGroups(t.DB.Group)
				t.GroupWidget.Select(t.GroupWidget.GetSelection())
			} else {
				t.Notify("Press d again to delete this feed.", false)
				t.ConfirmationStatus = 'd'
			}
		}
	case 'j':
		row, _ := t.GroupWidget.GetSelection()
		if row == t.GroupWidget.GetRowCount()-1 || t.GroupWidget.GetRowCount() == 0 {
			t.focusLeftTable(enumFeedWidget)
			return nil
		}
	case 'l':
		t.setFocus(t.ItemWidget.Box)
	case 'J':
		t.focusLeftTable(enumFeedWidget)
	}

	return event
}

func (t *Tui) feedTableInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {

	}

	switch event.Rune() {
	case 'c':
		t.ColorWidget.Clear()
		for i, c := range t.getColorRange() {
			t.ColorWidget.SetCell(i, 0,
				tview.NewTableCell(strconv.Itoa(c)).
					SetTextColor(tcell.Color(c+1<<32)))
		}

		t.Pages.ShowPage(colorTable)
		t.App.SetFocus(t.ColorWidget)
		t.Notify("Select a color to change the feed's one.", false)
		return nil
	case 'd':
		if t.IsLoading {
			t.Notify(msgRefusedByLoading, true)
		} else {
			if t.ConfirmationStatus == 'd' {
				cell := t.FeedWidget.GetCell(t.FeedWidget.GetSelection())
				ref, ok := cell.GetReference().(*FeedCellRef)
				if ok {
					if err := t.DB.DeleteFeed(ref.Feed); err != nil {
						panic(err)
					}
					t.Notify("deleted.", false)
				} else {
					t.Notify("delete failed.", true)
				}
				t.ConfirmationStatus = defaultConfirmationStatus
				t.resetGroups(t.DB.Group)
				t.resetFeeds(t.DB.Feed)
				t.FeedWidget.Select(t.FeedWidget.GetSelection())
			} else {
				t.Notify("Press d again to delete this feed.", false)
				t.ConfirmationStatus = 'd'
			}
		}
	case 'm':
		if t.IsLoading {
			t.Notify(msgRefusedByLoading, true)
		} else {
			if len(t.SelectingFeeds) == 0 {
				t.Notify("Select at least 1 Feed to make a Group.", true)
				return nil
			}
			t.InputWidget.SetTitle("New Group")
			t.InputWidget.Mode = 'm'
			t.Pages.ShowPage(inputField)
			t.setFocus(t.InputWidget.Box)
			t.Notify("Enter a new group title.", false)
			return nil
		}
	case 'v':
		cell := t.FeedWidget.Table.GetCell(t.FeedWidget.Table.GetSelection())
		cellRef, ok := cell.GetReference().(*FeedCellRef)
		feed := cellRef.Feed
		if ok {
			if cell.BackgroundColor == tview.Styles.PrimitiveBackgroundColor {
				cell.SetBackgroundColor(tview.Styles.PrimaryTextColor)
				t.SelectingFeeds = append(t.SelectingFeeds, feed)
			} else {
				cell.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
				for i, f := range t.SelectingFeeds {
					if f == feed {
						t.SelectingFeeds = append(t.SelectingFeeds[:i], t.SelectingFeeds[i+1:]...)
						break
					}
				}
			}
		}
	case 'k':
		row, _ := t.FeedWidget.GetSelection()
		if row == 0 {
			t.focusLeftTable(enumGroupWidget)
			return nil
		}
	case 'l':
		t.setFocus(t.ItemWidget.Box)
	case 'K':
		t.focusLeftTable(enumGroupWidget)
	}

	return event
}

func (t *Tui) itemTableInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {

	}

	switch event.Rune() {
	case 'c':
		t.ColorWidget.Clear()
		for i, c := range t.getColorRange() {
			t.ColorWidget.SetCell(i, 0,
				tview.NewTableCell(strconv.Itoa(c)).
					SetTextColor(tcell.Color(c+1<<32)))
		}

		t.Pages.ShowPage(colorTable)
		t.App.SetFocus(t.ColorWidget)
		t.Notify("Select a color to change the feed's one.", false)
		return nil
	case 'h':
		t.focusLeftTable(t.CurrentLeftTable)
	case 'o':
		row, _ := t.ItemWidget.GetSelection()
		item, err := t.ItemWidget.GetItem(row)
		if err != nil {
			panic(err)
		}
		if err := openURL(item.Link); err != nil {
			panic(err)
		}
		return nil
	}

	return event
}

func (t *Tui) inputWidgetInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		t.Pages.SwitchToPage(mainPage)
		t.focusLeftTable(t.CurrentLeftTable)
		t.InputWidget.SetText("")
		t.InputWidget.SetTitle("Input")
	case tcell.KeyEnter:
		switch t.InputWidget.Mode {
		case 'm':
			if err := t.MakeGroup(t.InputWidget.GetText()); err != nil {
				t.Notify(err.Error(), true)
			}
			for i := 0; i < t.FeedWidget.GetRowCount(); i++ {
				cell := t.FeedWidget.GetCell(i, 0)
				if cell.BackgroundColor == tview.Styles.PrimaryTextColor {
					cell.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
				}
			}
			t.SelectingFeeds = []*fd.Feed{}
		case 'n':
			if err := t.AddFeedFromURL(t.InputWidget.GetText()); err != nil {
				t.Notify(err.Error(), true)
			}
		}
		db.SortGroup(t.DB.Group)
		t.resetGroups(t.DB.Group)
		t.Pages.SwitchToPage(mainPage)
		t.focusLeftTable(t.CurrentLeftTable)
		t.InputWidget.SetText("")
		t.InputWidget.SetTitle("Input")
	}

	switch event.Rune() {
	}

	return event
}

func (t *Tui) descriptionWidgetInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'j', 'k':
		return event
	case 'D':
		return nil
	}
	t.Pages.HidePage(descriptionField)
	t.setFocus(t.LastFocusedWidget)
	return nil
}

func (t *Tui) colorWidgetInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		t.Pages.SwitchToPage(mainPage)
		t.focusLeftTable(t.CurrentLeftTable)
		t.ColorWidget.Clear()
	case tcell.KeyEnter:
		color, err := strconv.Atoi(t.ColorWidget.GetCell(t.ColorWidget.GetSelection()).Text)
		if err != nil {
			panic(err)
		}

		db.SortGroup(t.DB.Group)
		t.resetGroups(t.DB.Group)

		if t.LastFocusedWidget == t.FeedWidget.Box {
			cell := t.FeedWidget.GetCell(t.FeedWidget.GetSelection())
			ref, ok := cell.GetReference().(*FeedCellRef)
			if ok {
				ref.Feed.SetColor(color)
				t.FeedWidget.setCell(ref.Feed)
				if err := db.SaveFeed(ref.Feed); err != nil {
					panic(err)
				}
				t.Notify("recolored.", false)
				t.feedTableSelectionChangedFunc(t.FeedWidget.GetSelection())
			} else {
				t.Notify("recolor failed.", true)
			}

			t.Pages.SwitchToPage(mainPage)
			t.focusLeftTable(t.CurrentLeftTable)
		} else if t.LastFocusedWidget == t.ItemWidget.Box {
			cell := t.ItemWidget.GetCell(t.ItemWidget.GetSelection())
			item, ok := cell.GetReference().(*fd.Item)
			if ok {
				parentFeed := t.DB.GetItemParent(item)
				parentFeed.SetColor(color)
				t.FeedWidget.setCell(parentFeed)
				if err := db.SaveFeed(parentFeed); err != nil {
					panic(err)
				}

				// SelectionChangedFuncを発火して色の変更を反映する
				t.focusLeftTable(t.CurrentLeftTable)
				t.setFocus(t.ItemWidget.Box)

				t.Notify("recolored.", false)
			} else {
				t.Notify("recolor failed.", true)
			}

			t.Pages.SwitchToPage(mainPage)
			t.setFocus(t.ItemWidget.Box)
		} else {
			panic(errors.New("error while colorchanging"))
		}
	}

	switch event.Rune() {
	}

	return event
}
