package models

import "time"

type TweetData struct {
	Year     int
	Month    int
	Day      int
	Text     string
	Hour     int
	Minute   int
	Title    string
	URL      string
	Author   string
	Views    int
	Likes    int
	Comments int
	Created  time.Time
}

type InstagramData struct {
	Year      int
	Month     int
	Day       int
	Text      string
	Hour      int
	Minute    int
	AccountID string
	Title     string
	URL       string
	Author    string
	Likes     int
	Comments  int
	Created   time.Time
}
