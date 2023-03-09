package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	colorFocused   = tcell.ColorWhite
	colorUnfocused = tcell.ColorGray
)

func (t *Tui) setFocusFunc() {
	t.GroupWidget.SetFocusFunc(func() {
		t.highlightBox(t.GroupWidget.Box)
		t.groupTableSelectionChangedFunc(t.GroupWidget.GetSelection())
	})
	t.FeedWidget.SetFocusFunc(func() {
		t.highlightBox(t.FeedWidget.Box)
		t.feedTableSelectionChangedFunc(t.FeedWidget.GetSelection())
	})
	t.ItemWidget.SetFocusFunc(func() {
		t.highlightBox(t.ItemWidget.Box)
		t.itemTableSelectionChangedFunc(t.ItemWidget.GetSelection())
	})
	t.InputWidget.SetFocusFunc(func() {
		help := [][]string{
			{"Esc", "close InputWidget"},
		}
		t.Help(help)
	})
	t.DescriptionWidget.SetFocusFunc(func() {
		t.highlightBox(t.DescriptionWidget.Box)

		help := [][]string{
			{"h", "↑"},
			{"j", "↓"},
			{"else", "close"},
			{"\n", ""},
		}
		t.Help(append(help, t.commonKeyHelp()...))
	})
	t.ColorWidget.SetFocusFunc(func() {
		t.highlightBox(t.ColorWidget.Box)
		t.itemTableSelectionChangedFunc(t.ColorWidget.GetSelection())
	})
}

func (t *Tui) highlightBox(b *tview.Box) {
	b.SetBorderColor(colorFocused)
}

func (t *Tui) setBlurFunc() {
	t.GroupWidget.SetBlurFunc(func() {
		t.darkenBox(t.GroupWidget.Box)
		t.ConfirmationStatus = defaultConfirmationStatus
	})
	t.FeedWidget.SetBlurFunc(func() {
		t.darkenBox(t.FeedWidget.Box)
		t.ConfirmationStatus = defaultConfirmationStatus
	})
	t.ItemWidget.SetBlurFunc(func() {
		t.darkenBox(t.ItemWidget.Box)
		t.ConfirmationStatus = defaultConfirmationStatus
	})
	t.DescriptionWidget.SetBlurFunc(func() {
		t.darkenBox(t.DescriptionWidget.Box)
		t.ConfirmationStatus = defaultConfirmationStatus
	})
	t.ColorWidget.SetFocusFunc(func() {
		t.darkenBox(t.ColorWidget.Box)
		t.ConfirmationStatus = defaultConfirmationStatus
	})
}

func (t *Tui) darkenBox(b *tview.Box) {
	b.SetBorderColor(colorUnfocused)
}
