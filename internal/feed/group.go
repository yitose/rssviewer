package feed

import (
	"bytes"
	"encoding/gob"
)

type Group struct {
	Title          string
	IsFirstUpdated bool
	FeedLinks      []string
}

func MergeFeeds(feeds []*Feed, title string) *Group {
	mergedFeedlinks := []string{}
	for _, feed := range feeds {
		mergedFeedlinks = append(mergedFeedlinks, feed.FeedLink)
	}
	resultFeed := &Group{
		Title:          title,
		IsFirstUpdated: false,
		FeedLinks:      mergedFeedlinks,
	}
	return resultFeed
}

func EncodeGroup(g *Group) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(g)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeGroup(data []byte) *Group {
	var g Group
	buf := bytes.NewBuffer(data)
	_ = gob.NewDecoder(buf).Decode(&g)
	return &g
}
