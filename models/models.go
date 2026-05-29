package models

import "time"

type TweetData struct {
	Year    int
	Month   int
	Day     int
	Text    string
	Hour    int
	Minute  int
	Title   string
	URL     string
	Author  string
	Views   int
	Likes   int
	Created time.Time
}
