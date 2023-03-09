package tui

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/rivo/tview"
	"github.com/yitose/rssviewer/internal/color"
	db "github.com/yitose/rssviewer/internal/db"
	fd "github.com/yitose/rssviewer/internal/feed"
	"github.com/yitose/rssviewer/pkg/util"
)

type TuiInterface interface {
	SortFeed()
	SelectFeed()
	Notify(text string)
	Descript(info [][]string)
}

type Tui struct {
	Config             *db.Config
	DB                 *db.FeedDB
	App                *tview.Application
	Pages              *tview.Pages
	GroupWidget        *GroupTable
	FeedWidget         *FeedTable
	ItemWidget         *ItemTable
	DescriptionWidget  *tview.TextView
	InfoWidget         *tview.TextView
	HelpWidget         *tview.TextView
	InputWidget        *InputBox
	ColorWidget        *tview.Table
	SelectingFeeds     []*fd.Feed
	LastFocusedWidget  *tview.Box
	ConfirmationStatus rune
	CurrentLeftTable   int
	IsLoading          bool
}

const (
	descriptionField          = "descPopup"
	inputField                = "InputPopup"
	colorTable                = "ColorTablePopup"
	mainPage                  = "MainPage"
	keymapPage                = "KeymapPage"
	defaultConfirmationStatus = '0'
	groupWidgetTitle          = "Groups"
	FeedWidgetTitle           = "Feeds"
	itemWidgetTitle           = "Items"
	descriptionWidgetTitle    = "Description"
	infoWidgetTitle           = "Info"
	colorWidgetTitle          = "Color"
)

const (
	enumGroupWidget = iota
	enumFeedWidget
)

var ErrImportFileNotFound = errors.Errorf(db.ImportListPath + " not found")

func NewTui() *Tui {
	tview.Styles.ContrastBackgroundColor = tview.Styles.PrimitiveBackgroundColor

	tui := &Tui{
		Config:             db.LoadOrNewConfig(),
		DB:                 db.NewDB(),
		App:                tview.NewApplication(),
		Pages:              tview.NewPages(),
		GroupWidget:        &GroupTable{FeedTable: &FeedTable{Table: newTable(groupWidgetTitle)}},
		FeedWidget:         &FeedTable{Table: newTable(FeedWidgetTitle)},
		ItemWidget:         &ItemTable{Table: newTable(itemWidgetTitle)},
		DescriptionWidget:  newTextView(descriptionWidgetTitle),
		InfoWidget:         newTextView(infoWidgetTitle),
		HelpWidget:         tview.NewTextView().SetTextAlign(1).SetDynamicColors(true),
		InputWidget:        &InputBox{InputField: newInputField(), Mode: 0},
		ColorWidget:        newTable(colorWidgetTitle),
		SelectingFeeds:     []*fd.Feed{},
		LastFocusedWidget:  nil,
		ConfirmationStatus: defaultConfirmationStatus,
		CurrentLeftTable:   enumGroupWidget,
		IsLoading:          false,
	}

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(tui.GroupWidget, 0, 2, false).
				AddItem(tui.FeedWidget, 0, 2, false).
				AddItem(tui.InfoWidget, 0, 1, false),
				0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(tui.ItemWidget, 0, 3, false).
				AddItem(tui.DescriptionWidget, 0, 1, false),
				0, 2, false),
			0, 1, false).AddItem(tui.HelpWidget, 2, 0, false)

	inputFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tui.InputWidget, 3, 1, false).
			AddItem(nil, 0, 1, false), 40, 1, false).
		AddItem(nil, 0, 1, false)

	descriptionFlex := tview.NewFlex().
		AddItem(nil, 0, 2, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tui.DescriptionWidget, 0, 8, false).
			AddItem(nil, 0, 1, false),
			0, 6, false).
		AddItem(nil, 0, 2, false)

	colorTableFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tui.ColorWidget, 0, 5, false).
			AddItem(nil, 0, 1, false), 40, 1, false).
		AddItem(nil, 0, 1, false)

	tui.Pages.
		AddPage(mainPage, mainFlex, true, true).
		AddPage(inputField, inputFlex, true, false).
		AddPage(colorTable, colorTableFlex, true, false).
		AddPage(descriptionField, descriptionFlex, true, false)

	tui.App.SetRoot(tui.Pages, true)

	tui.setKeyBinding()
	tui.setSelectionFunc()
	tui.setFocusFunc()
	tui.setBlurFunc()

	return tui
}

func (t *Tui) setFocus(p *tview.Box) {
	t.LastFocusedWidget = p
	t.App.SetFocus(p)
}

func (t *Tui) getRandomColor() int {
	maxHue := t.Config.Color.MaxHue
	minHue := t.Config.Color.MinHue
	maxSaturatio := t.Config.Color.MaxSaturatio
	minSaturatio := t.Config.Color.MinSaturatio
	maxLightness := t.Config.Color.MaxLightness
	minLightness := t.Config.Color.MinLightness
	return color.GetRandomColor(maxHue, minHue, maxSaturatio, minSaturatio, maxLightness, minLightness)
}

func (t *Tui) getColorRange() []int {
	maxHue := t.Config.Color.MaxHue
	minHue := t.Config.Color.MinHue
	maxSaturatio := t.Config.Color.MaxSaturatio
	minSaturatio := t.Config.Color.MinSaturatio
	maxLightness := t.Config.Color.MaxLightness
	minLightness := t.Config.Color.MinLightness
	return color.GetColorRange(maxHue, minHue, maxSaturatio, minSaturatio, maxLightness, minLightness)
}

func (t *Tui) MakeGroup(title string) error {
	if len(t.SelectingFeeds) == 0 {
		return nil
	}
	if err := t.DB.AddOrUpdateGroup(fd.MergeFeeds(t.SelectingFeeds, title)); err != nil {
		return err
	}

	return nil
}

func (t *Tui) Descript(desc [][]string) {
	var s string
	for _, line := range desc {
		s += fmt.Sprint("[#a0a0a0::b]", line[0], "[-::-] ", line[1], "\n")
	}
	t.DescriptionWidget.SetText(s).ScrollToBeginning()
}

func (t *Tui) Notify(m string, red bool) {
	if red {
		m = "[#ff0000::b]" + m
	}
	t.InfoWidget.SetText(m)
}

func (t *Tui) Help(help [][]string) {
	var s string
	for _, line := range help {
		if line[0] == "\n" {
			s += "\n"
		} else {
			s += fmt.Sprint("[-::-][", line[0], "[][#a0a0a0::b] ", line[1], " ")
		}
	}
	t.HelpWidget.SetText(s)
}

func (t *Tui) AddFeedFromURL(url string) error {
	for _, f := range t.DB.Feed {
		if f.FeedLink == url {
			return nil
		}
	}

	newFeed, err := fd.GetFeedFromURL(url, t.getRandomColor())
	if err != nil {
		t.Notify(err.Error(), true)
		return nil
	}

	t.DB.Feed = append(t.DB.Feed, newFeed)

	if err := db.SaveFeed(newFeed); err != nil {
		return err
	}

	db.SortFeed(t.DB.Feed)
	t.resetFeeds(t.DB.Feed)

	return nil
}

func (t *Tui) AddFeedsFromURL(path string) error {
	if !util.IsFile(path) {
		return ErrImportFileNotFound
	}

	_, feedURLs, err := util.GetLines(path)
	if err != nil {
		return err
	}

	f := func(url string, done chan<- bool) {
		if err := t.AddFeedFromURL(url); err != nil {
			panic(err)
		}
		done <- true
	}

	n := len(feedURLs)

	t.IsLoading = true

	done := make(chan bool, n)
	for _, url := range feedURLs {
		go f(url, done)
	}

	c := 0
	for i := 0; i < n; i++ {
		t.Notify(fmt.Sprintf("Importing Feeds...(%d/%d)", c, n), false)
		t.App.Draw()
		<-done
		c++
	}

	t.IsLoading = false

	return nil
}

func (t *Tui) UpdateAllFeed() error {
	n := len(t.DB.Feed)

	f := func(url string, done chan<- *fd.Feed) {
		feed, _ := fd.GetFeedFromURL(url, 0)
		done <- feed
	}

	t.IsLoading = true

	done := make(chan *fd.Feed, n)
	for _, feed := range t.DB.Feed {
		go f(feed.FeedLink, done)
	}

	isLoadedFeedList := map[string]int{}
	loadedFeeds := []*fd.Feed{}
	isLoadedGroupList := map[string]bool{}
	loadedGroups := []*fd.Group{}

	c := 0
	for i := 0; i < n; i++ {
		f := <-done
		isLoadedFeedList[f.FeedLink] = 1
		for i, feed := range t.DB.Feed {
			if feed.FeedLink == f.FeedLink {
				c++
				f.SetColor(feed.Color)
				t.DB.Feed[i] = f
				loadedFeeds = append(loadedFeeds, f)
				t.Notify(fmt.Sprintf("Updating Feeds...(%d/%d)", c, n), false)
			}
		}

		for _, g := range t.DB.Group {
			loadedUrlCount := 0
			for _, link := range g.FeedLinks {
				loadedUrlCount += isLoadedFeedList[link]
			}
			if loadedUrlCount == len(g.FeedLinks) && !isLoadedGroupList[g.Title] {
				loadedGroups = append(loadedGroups, g)
				isLoadedGroupList[g.Title] = true
			}
		}

		db.SortFeed(loadedFeeds)
		t.resetFeeds(loadedFeeds)
		db.SortGroup(loadedGroups)
		t.resetGroups(loadedGroups)

		wasFocusItemWidget := t.ItemWidget.HasFocus()
		t.focusLeftTable(t.CurrentLeftTable)
		if wasFocusItemWidget {
			t.setFocus(t.ItemWidget.Box)
		}

		t.App.Draw()
	}

	t.IsLoading = false

	db.SortGroup(t.DB.Group)
	t.resetGroups(t.DB.Group)
	db.SortFeed(t.DB.Feed)
	t.resetFeeds(t.DB.Feed)

	if t.ItemWidget.HasFocus() {
		t.focusLeftTable(t.CurrentLeftTable)
		t.setFocus(t.ItemWidget.Box)
	}

	if t.FeedWidget.GetRowCount() > 0 {
		t.Notify("All feeds are up to date.", false)
	} else {
		t.Notify("Hello User! Press [n[] to add the first feed.", false)
	}

	return nil
}

func (t *Tui) focusLeftTable(enum int) {
	if enum != enumGroupWidget && enum != enumFeedWidget {
		return
	}

	switch enum {
	case enumGroupWidget:
		t.setFocus(t.GroupWidget.Box)
	case enumFeedWidget:
		t.setFocus(t.FeedWidget.Box)
	}

	t.CurrentLeftTable = enum
}

func (t *Tui) resetFeeds(feeds []*fd.Feed) {
	if len(feeds) == 0 {
		return
	}

	feedCellRefList := map[string]*FeedCellRef{}
	for i := 0; i < t.FeedWidget.GetRowCount(); i++ {
		cell := t.FeedWidget.GetCell(i, 0)
		ref, ok := cell.GetReference().(*FeedCellRef)
		if ok {
			feedCellRefList[cell.Text] = ref
		}
	}

	t.FeedWidget.Clear()

	for _, f := range feeds {
		cell := t.FeedWidget.setCell(f)
		if cellRef, ok := feedCellRefList[cell.Text]; ok {
			cell.SetReference(cellRef)
		}
	}
}

func (t *Tui) resetGroups(groups []*fd.Group) {
	if len(groups) == 0 {
		return
	}

	groupCellRefList := map[string]*GroupCellRef{}
	for i := 0; i < t.GroupWidget.GetRowCount(); i++ {
		cell := t.GroupWidget.GetCell(i, 0)
		ref, ok := cell.GetReference().(*GroupCellRef)
		if ok {
			groupCellRefList[cell.Text] = ref
		}
	}

	t.GroupWidget.Clear()

	if groups[0].Title != db.TodaysFeedTitle {
		groups = append([]*fd.Group{{Title: db.TodaysFeedTitle}}, groups...)
	}

	for _, g := range groups {
		cell := t.GroupWidget.setCell(g)
		if cellRef, ok := groupCellRefList[cell.Text]; ok {
			cell.SetReference(cellRef)
		}
	}
}

func (t *Tui) Run() error {

	if err := t.DB.LoadFeeds(); err != nil {
		return err
	}

	if len(t.DB.Group) > 0 {
		t.setFocus(t.GroupWidget.Table.Box)
	} else {
		t.setFocus(t.FeedWidget.Table.Box)
	}

	go func() {
		if err := t.UpdateAllFeed(); err != nil {
			panic(err)
		}

		t.App.QueueUpdateDraw(func() {})
	}()

	if err := t.App.Run(); err != nil {
		t.App.Stop()
		return err
	}

	return nil
}
