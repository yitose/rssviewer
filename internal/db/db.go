package db

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	fd "github.com/yitose/rssviewer/internal/feed"
	"github.com/yitose/rssviewer/pkg/util"
)

const (
	dataRoot        = "rssviewer"
	TodaysFeedTitle = "Today's Articles"
	SavePrefixGroup = "g_"
	SavePrefixFeed  = "f_"
)

var (
	DataPath       = filepath.Join(getDataPath(), "data")
	ExportListPath = filepath.Join(getDataPath(), "list_export.txt")
	ImportListPath = filepath.Join(getDataPath(), "list.txt")
	ConfigPath     = filepath.Join(getDataPath(), "config.json")
)

type DBInterface interface {
	LoadFeed(path string) error
	UpdateAllFeed() error
	AddOrUpdateGroup(g *fd.Feed)
}

type FeedDB struct {
	Group []*fd.Group
	Feed  []*fd.Feed
}

func NewDB() *FeedDB {
	db := &FeedDB{
		Group: []*fd.Group{},
		Feed:  []*fd.Feed{},
	}
	return db
}

func getDataPath() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, dataRoot)
}

func (d *FeedDB) LoadFeeds() error {
	if !util.IsDir(DataPath) {
		if err := os.MkdirAll(DataPath, 0755); err != nil {
			return err
		}
	}

	for _, file := range util.DirWalk(DataPath) {
		b, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if strings.HasPrefix(filepath.Base(file), SavePrefixGroup) {
			d.Group = append(d.Group, fd.DecodeGroup(b))
		} else {
			d.Feed = append(d.Feed, fd.DecodeFeed(b))
		}
	}

	SortFeed(d.Feed)

	return nil
}

func (d *FeedDB) AddOrUpdateGroup(g *fd.Group) error {
	var sameNameGroup *fd.Group

	sameNameGroup = nil
	for _, group := range d.Group {
		if g.Title == group.Title {
			sameNameGroup = group
			break
		}
	}

	if sameNameGroup == nil {
		// Add g
		d.Group = append(d.Group, g)
		if err := SaveGroup(g); err != nil {
			return err
		}
	} else {
		// Update sameNameGroup
		for _, l := range g.FeedLinks {
			isNewFeedLink := true
			for _, url := range sameNameGroup.FeedLinks {
				if l == url {
					isNewFeedLink = false
					break
				}
			}
			if isNewFeedLink {
				sameNameGroup.FeedLinks = append(sameNameGroup.FeedLinks, l)
			}
		}
		if err := SaveGroup(sameNameGroup); err != nil {
			return err
		}
	}

	SortGroup(d.Group)

	return nil
}

func SaveGroup(g *fd.Group) error {
	if g.Title == TodaysFeedTitle {
		return nil
	}
	b, err := fd.EncodeGroup(g)
	if err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(g.Title)))
	if err := util.SaveBytes(b, filepath.Join(DataPath, SavePrefixGroup+hash)); err != nil {
		return err
	}
	return nil
}

func SaveFeed(f *fd.Feed) error {
	copy := &fd.Feed{}

	*copy = *f
	copy.Items = nil

	b, err := fd.EncodeFeed(f)
	if err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(copy.FeedLink)))
	if err := util.SaveBytes(b, filepath.Join(DataPath, SavePrefixFeed+hash)); err != nil {
		return err
	}
	return nil
}

func SortFeed(feeds []*fd.Feed) {
	sort.Slice(feeds, func(i, j int) bool {
		return strings.Compare(feeds[i].Title, feeds[j].Title) == -1
	})
}

func SortGroup(groups []*fd.Group) {
	sort.Slice(groups, func(i, j int) bool {
		return strings.Compare(groups[i].Title, groups[j].Title) == -1
	})
}

func (d *FeedDB) DeleteGroup(g *fd.Group) error {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(g.Title)))
	if err := os.Remove(filepath.Join(DataPath, SavePrefixGroup+hash)); err != nil {
		return err
	}

	for i, group := range d.Group {
		if g.Title == group.Title {
			d.Group = append(d.Group[:i], d.Group[i+1:]...)
			break
		}
	}

	return nil
}

func (d *FeedDB) DeleteFeed(f *fd.Feed) error {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(f.FeedLink)))
	if err := os.Remove(filepath.Join(DataPath, SavePrefixFeed+hash)); err != nil {
		return err
	}

	for i, feed := range d.Feed {
		if f.FeedLink == feed.FeedLink {
			d.Feed = append(d.Feed[:i], d.Feed[i+1:]...)
			break
		}
	}

	for i, group := range d.Group {
		for j, link := range group.FeedLinks {
			if f.FeedLink == link {
				group.FeedLinks = append(group.FeedLinks[:j], group.FeedLinks[j+1:]...)
			}
			if len(group.FeedLinks) == 0 {
				d.Group = append(d.Group[:i], d.Group[i+1:]...)
			}
		}
	}

	return nil
}

func (d *FeedDB) GetItemParent(i *fd.Item) *fd.Feed {
	for _, f := range d.Feed {
		if f.FeedLink == i.Belong {
			return f
		}
	}
	return &fd.Feed{}
}
