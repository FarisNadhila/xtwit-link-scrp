package models

import "time"

type SocialData struct {
	Year      uint16
	Month     uint8
	Day       uint8
	Text      string
	Hour      uint8
	Minute    uint8
	Title     string
	URL       string
	Author    string
	Views     uint64
	Likes     uint32
	Comments  uint32
	Repost    uint32
	Quotes    uint32
	SumRepost uint32
	Share     uint32
	Following uint16
	Followers uint32
	Created   time.Time
}
