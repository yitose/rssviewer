package tui

import fd "github.com/yitose/rssviewer/internal/feed"

type FeedCellRef struct {
	Feed   *fd.Feed
	Cursor int
}

func NewFeedCellRef(f *fd.Feed) *FeedCellRef {
	return &FeedCellRef{
		Feed:   f,
		Cursor: 0,
	}
}

type GroupCellRef struct {
	Group  *fd.Group
	Cursor int
}

func NewGroupCellRef(g *fd.Group) *GroupCellRef {
	return &GroupCellRef{
		Group:  g,
		Cursor: 0,
	}
}
