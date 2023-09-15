package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

type Feed struct {
	URL   string     `json:"url"`
	Read  bool       `json:"read"`
	Items []FeedItem `json:"items"`
}

type FeedItem struct {
	Title string `json:"title"`
	Link  string `json:"link"`
	Read  bool   `json:"read"`
}

var (
	feedMap      = make(map[string]Feed)
	feedMapMutex = &sync.RWMutex{}
)

// parseAndAddFeed 會解析 RSS 並加入到 feedMap
func parseAndAddFeed(url string) (Feed, error) {
	fp := gofeed.NewParser()
	parsedFeed, err := fp.ParseURL(url)
	if err != nil {
		return Feed{}, err
	}

	feed := Feed{URL: url, Read: false}
	feed.Items = make([]FeedItem, len(parsedFeed.Items))
	for i, item := range parsedFeed.Items {
		feed.Items[i] = FeedItem{Title: item.Title, Link: item.Link, Read: false}
	}

	feedMapMutex.Lock()
	feedMap[url] = feed
	feedMapMutex.Unlock()

	return feed, nil
}

func subscribeFeed(w http.ResponseWriter, r *http.Request) {
	var feed Feed
	err := json.NewDecoder(r.Body).Decode(&feed)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	newFeed, err := parseAndAddFeed(feed.URL)
	if err != nil {
		http.Error(w, "Invalid RSS URL", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(newFeed)
}

func listSubscribedFeeds(w http.ResponseWriter, r *http.Request) {
	feedMapMutex.RLock()
	defer feedMapMutex.RUnlock()

	feeds := make([]Feed, 0, len(feedMap))
	for _, feed := range feedMap {
		feeds = append(feeds, feed)
	}

	json.NewEncoder(w).Encode(feeds)
}

func deleteFeed(w http.ResponseWriter, r *http.Request) {
	var feed Feed
	err := json.NewDecoder(r.Body).Decode(&feed)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	feedMapMutex.Lock()
	delete(feedMap, feed.URL)
	feedMapMutex.Unlock()

	// Return an empty feed to indicate success
	json.NewEncoder(w).Encode(Feed{})
}

func markItemRead(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		Read  bool   `json:"read"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	feedMapMutex.Lock()
	defer feedMapMutex.Unlock()

	feed, ok := feedMap[payload.URL]
	if !ok {
		http.Error(w, "Feed not found", http.StatusBadRequest)
		return
	}

	for i, item := range feed.Items {
		if item.Title == payload.Title {
			feed.Items[i].Read = payload.Read
			break
		}
	}

	feedMap[payload.URL] = feed

	// Return the updated feed to indicate success
	json.NewEncoder(w).Encode(feed)
}

func updateFeeds() {
	fp := gofeed.NewParser()
	feedMapMutex.Lock()
	for url, feed := range feedMap {
		parsedFeed, err := fp.ParseURL(url)
		if err != nil {
			fmt.Printf("Error parsing feed %s: %v\n", url, err)
			continue
		}

		existingItems := make(map[string]bool)
		for _, item := range feed.Items {
			existingItems[item.Title] = true
		}

		for _, newItem := range parsedFeed.Items {
			if _, exists := existingItems[newItem.Title]; !exists {
				feed.Items = append(feed.Items, FeedItem{Title: newItem.Title, Link: newItem.Link, Read: false})
			}
		}

		feedMap[url] = feed
	}
	feedMapMutex.Unlock()
}

func autoUpdateFeeds() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		updateFeeds()
	}
}

func main() {
	http.HandleFunc("/subscribe", subscribeFeed)
	http.HandleFunc("/list", listSubscribedFeeds)
	http.HandleFunc("/delete", deleteFeed)
	http.HandleFunc("/markRead", markItemRead)

	go autoUpdateFeeds()

	http.ListenAndServe(":8080", nil)
}
