package models

type Follower struct {
	UserId     int64 `db:"user_id"`
	FollowerId int64 `db:"follower_id"`
}
