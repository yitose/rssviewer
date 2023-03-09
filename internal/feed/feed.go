package feed

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
)

var (
	ErrUrlFailed   = "Parsing URL Failed: "
	ErrCmdFailed   = "Executing Command Failed: "
	ErrParseFailed = "Parseing Feed Failed: "
)

type Feed struct {
	*gofeed.Feed
	Color int
	Items []*Item
}

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func GetFeedFromURL(url string, color int) (*Feed, error) {
	var (
		parsedFeed *gofeed.Feed
		feed       *Feed
		err        error
	)
	parser := gofeed.NewParser()

	failureFeed := &Feed{
		Feed: &gofeed.Feed{
			Title:    "Error",
			FeedLink: url,
		},
		Color: int(tcell.ColorRed),
		Items: []*Item{},
	}

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	if isUrl(url) {
		parsedFeed, err = parser.ParseURL(url)
		if err != nil {
			errMsg := ErrUrlFailed + err.Error()
			failureFeed.Feed.Description = errMsg
			return failureFeed, errors.Errorf(errMsg)
		}
	} else {
		cmd := Cmd{Cmd: "sh", Args: []string{"-c"}}
		if runtime.GOOS == "windows" {
			cmd = Cmd{Cmd: "powershell.exe", Args: []string{"-Command"}}
		}

		output, err := exec.Command(cmd.Cmd, append(cmd.Args, url)...).Output()
		if err != nil {
			errMsg := ErrCmdFailed + err.Error()
			failureFeed.Feed.Description = errMsg
			return failureFeed, errors.Errorf(ErrCmdFailed + err.Error())
		}
		parsedFeed, err = parser.ParseString(string(output))
		if err != nil {
			errMsg := ErrParseFailed + err.Error() + string(output)
			failureFeed.Feed.Description = errMsg

			if err := os.WriteFile(filepath.Join(home, "fd.log"), output, 0755); err != nil {
				panic(err)
			}

			return failureFeed, errors.Errorf(ErrParseFailed + err.Error())
		}
	}

	parsedFeed.FeedLink = url

	rawItems := []*gofeed.Item{}
	for i := 0; i < len(parsedFeed.Items); i++ {
		rawItems = append(rawItems, &gofeed.Item{})
	}

	for i, t := range parsedFeed.Items {
		*rawItems[i] = *t
	}

	feed = &Feed{
		Feed:  parsedFeed,
		Color: color,
		Items: []*Item{},
	}

	for _, item := range rawItems {
		if item.PublishedParsed != nil && time.Now().After(*item.PublishedParsed) {
			jstTime := item.PublishedParsed.In(time.Local)
			item.PublishedParsed = &jstTime
			feed.Items = append(feed.Items, &Item{
				Item:   item,
				Belong: feed.FeedLink,
				Color:  feed.Color,
			})
		}
	}

	SortItems(feed.Items)
	return feed, nil
}

func SortItems(items []*Item) {
	sort.Slice(items, func(i, j int) bool {
		a := items[i]
		b := items[j]

		if a.PublishedParsed.Equal(*b.PublishedParsed) {
			return strings.Compare(a.Title, b.Title) == 1
		} else {
			return a.PublishedParsed.After(*b.PublishedParsed)
		}
	})
}

func (f *Feed) SetColor(color int) {
	f.Color = color
	for _, item := range f.Items {
		item.Color = color
	}
}

type Cmd struct {
	Cmd  string
	Args []string
}
