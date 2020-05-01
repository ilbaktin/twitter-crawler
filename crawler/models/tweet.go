package models

type Tweet struct {
	CreatedAt            string  `json:"created_at"`
	FullText             string  `json:"full_text"`
	IdStr                string  `json:"id_str"`
	Lang                 string  `json:"lang"`
	ReplyCount           float64 `json:"reply_count"`
	RetweetCount         float64 `json:"retweet_count"`
	RetweetedStatusIdStr string  `json:"retweeted_status_id_str"`
	UserIdStr            string  `json:"user_id_str"`
}
