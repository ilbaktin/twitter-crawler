package crawler_tasks

import "univer/twitter-crawler/models"

type usersJsonResponse struct {
	Users             []*models.User `json:"users"`
	NextCursor        float64        `json:"next_cursor"`
	NextCursorStr     string         `json:"next_cursor_str"`
	PreviousCursor    float64        `json:"previous_cursor"`
	PreviousCursorStr string         `json:"previous_cursor_str"`
}

type tweetJsonResponse struct {
	GlobalObjects globalObjects `json:"globalObjects"`
}

type globalObjects struct {
	Tweets map[string]*models.Tweet `json:"tweets"`
	Users  map[string]*models.User  `json:"users"`
}
