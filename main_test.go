package main

import (
	"testing"
)

func TestParseAndAddFeed(t *testing.T) {
	// 測試提供無效 RSS URL 的情況
	t.Run("invalid RSS URL", func(t *testing.T) {
		_, err := parseAndAddFeed("invalid_rss_url")
		if err == nil {
			t.Error("Expected an error for invalid RSS URL, got none")
		}
	})

	// 測試成功訂閱一個 RSS feed
	t.Run("valid RSS URL", func(t *testing.T) {
		// 使用一個預先知道的，有效的 RSS URL 進行測試
		validRSSURL := "https://lorem-rss.herokuapp.com/feed"
		feed, err := parseAndAddFeed(validRSSURL)

		if err != nil {
			t.Errorf("Didn't expect an error, got %v", err)
		}

		if feed.URL != validRSSURL {
			t.Errorf("Expected feed URL to be %s, got %s instead", validRSSURL, feed.URL)
		}

		if len(feed.Items) == 0 {
			t.Errorf("Expected feed items to be populated, got an empty list")
		}
	})
}
