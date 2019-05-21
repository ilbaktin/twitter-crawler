package crawler_tasks

import "univer/twitter-crawler/models"

type usersJsonResponseNew struct {
	Users				[]*models.User	`json:"users"`
	NextCursor			float64			`json:"next_cursor"`
	NextCursorStr		string			`json:"next_cursor_str"`
	PreviousCursor		float64			`json:"previous_cursor"`
	PreviousCursorStr	string			`json:"previous_cursor_str"`
}



type usersJsonResponse struct {
	Users				[]*userJson		`json:"users"`
	NextCursor			float64			`json:"next_cursor"`
	NextCursorStr		string			`json:"next_cursor_str"`
	PreviousCursor		float64			`json:"previous_cursor"`
	PreviousCursorStr	string			`json:"previous_cursor_str"`
}

type userJson struct {
	Id				float64		`json:"id"`
	IdStr			string		`json:"id_str"`
	Name			string		`json:"name"`
	ScreenName		string		`json:"screen_name"`
	FollowersCount	float64		`json:"followers_count"`
	FriendsCount	float64		`json:"friends_count"`
	CreatedAt		string		`json:"created_at"`
}

type showJsonResponse struct {
	Id				float64		`json:"id"`
	IdStr			string		`json:"id_str"`
	Name			string		`json:"name"`
	ScreenName		string		`json:"screen_name"`
	Location		string		`json:"location"`
	ProfileLocation	string		`json:"profile_location"`
	Description		string		`json:"description"`
	Url				string		`json:"url"`
	FollowersCount	float64		`json:"followers_count"`
	FriendsCount	float64		`json:"friends_count"`
	CreatedAt		string		`json:"created_at"`
}





type usersJsonResponseOld struct {
	MinPosition		string		`json:"min_position"`
	HasMoreItems	bool		`json:"has_more_items"`
	ItemsHtml		string		`json:"items_html"`
	NewLatentCount	float64		`json:"new_latent_count"`
}

type errorJsonResponseOld struct {
	Message			*string		`json:"message"`
}
