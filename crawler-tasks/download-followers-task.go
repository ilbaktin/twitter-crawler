package crawler_tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"univer/twitter-crawler/conf"
	"univer/twitter-crawler/log"
	"univer/twitter-crawler/models"
	"univer/twitter-crawler/storage"
)

type DownloadFollowersTask struct {
	ScreenName		string

	cookies			map[string]string
	headers			map[string]string
	*log.Logger
}

func (task DownloadFollowersTask) Exec(stor storage.Storage) error {
	task.Logger = log.NewLogger(fmt.Sprintf("DownloadFollowersTask '%s'", task.ScreenName))

	config, err := conf.LoadConfig()
	if err != nil {
		return fmt.Errorf("can't load config, err='%v'", err)
	}
	if !checkRequiredCookiesExists(config.Cookies) {
		return fmt.Errorf("not all of required cookies are set in config, should be set %v", requiredCookies)
	}
	if !checkRequiredHeadersExists(config.Headers) {
		return fmt.Errorf("not all of required headers are set in config, should be set %v", requiredHeaders)
	}
	task.cookies = config.Cookies
	task.headers = config.Headers

	user, err := stor.GetUserByScreenName(task.ScreenName)
	if err != nil {
		task.LogInfo(fmt.Sprintf("User '%s' not found in db, requesting...", task.ScreenName))
		user, err = task.doShowRequest()
		if err != nil {
			return err
		}
		err = stor.AddNewUsers([]*models.User{user})
		if err != nil {
			return err
		}
	}
	cursor := user.NextCursorStr

	for cursor != "0" {
		usersJsonResp, err := task.doUsersRequest(user.Id, cursor)
		if err != nil {
			return err
		}
		users := usersJsonResp.Users
		followers := make([]*models.Follower, 0, len(users))
		for _, follower := range users {
			followers = append(followers, &models.Follower{
				UserId: user.Id,
				FollowerId: follower.Id,
			})
		}
		err = stor.AddNewUsers(users)
		if err != nil {
			return err
		}
		err = stor.AddNewFollowers(followers)
		if err != nil {
			return err
		}

		user.NextCursor = int64(usersJsonResp.NextCursor)
		user.NextCursorStr = usersJsonResp.NextCursorStr
		user.IsFollowersDownloaded = usersJsonResp.NextCursor == 0
		err = stor.UpdateUserState(user)
		if err != nil {
			return err
		}

		cursor = user.NextCursorStr
	}
	task.LogInfo(fmt.Sprintf("Followers successfully downloaded."))

	return nil
}

func (task DownloadFollowersTask) doShowRequest() (user *models.User, err error)  {
	showReq, err := task.createShowRequest()
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(showReq)
	if err != nil {
		return nil, fmt.Errorf("request failed, err='%v'", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad response status code, got %d, want 200", resp.StatusCode)
	}

	jsonBytes, err := ioutil.ReadAll(resp.Body)

	user = &models.User{}
	err = json.Unmarshal(jsonBytes, user)
	if err != nil {
		return nil, err
	}

	//user.AdditionalData = string(jsonBytes)
	user.NextCursor = -1
	user.NextCursorStr = "-1"

	return user, nil
}

func (task DownloadFollowersTask) doOptionsUsersRequest(userId int64, cursor string) error {
	optionsReq, err := task.createUsersOptionsRequest(userId, cursor)
	if err != nil {
		return err
	}
	optionsResp, err := http.DefaultClient.Do(optionsReq)
	if err != nil {
		return errors.New(fmt.Sprintf("options users request failed, err='%v'", err))
	}
	defer optionsResp.Body.Close()

	return nil
}

func (task DownloadFollowersTask) doShowOptionsRequest() error {
	optionsReq, err := task.createShowOptionsRequest()
	if err != nil {
		return err
	}
	optionsResp, err := http.DefaultClient.Do(optionsReq)
	if err != nil {
		return errors.New(fmt.Sprintf("options show request failed, err='%v'", err))
	}
	defer optionsResp.Body.Close()

	return nil
}

func (task DownloadFollowersTask) doUsersRequest(userId int64, cursor string) (userJsonResp *usersJsonResponse, err error) {
	usersReq, err := task.createUsersRequest(userId, cursor)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(usersReq)
	if err != nil {
		return nil, fmt.Errorf("request failed, err='%v'", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad response status code, got %d, want 200", resp.StatusCode)
	}

	jsonBytes, err := ioutil.ReadAll(resp.Body)

	usersJsonResp := &usersJsonResponse{}
	err = json.Unmarshal(jsonBytes, usersJsonResp)
	if err != nil {
		return nil, err
	}

	return usersJsonResp, nil

}

func (task DownloadFollowersTask) createShowOptionsRequest() (*http.Request, error) {
	req, err := http.NewRequest("OPTIONS","https://api.twitter.com/1.1/users/show.json", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()

	q.Add("include_profile_interstitial_type", "1")
	q.Add("include_blocking", "1")
	q.Add("include_blocked_by", "1")
	q.Add("include_followed_by", "1")
	q.Add("include_want_retweets", "1")
	q.Add("include_mute_edge", "1")
	q.Add("include_can_dm", "1")
	q.Add("include_can_media_tag", "1")
	q.Add("skip_status", "1")
	q.Add("screen_name", task.ScreenName)

	req.URL.RawQuery = q.Encode()


	req.Header.Set("Access-Control-Request-Headers", "authorization,x-csrf-token,x-twitter-active-user,x-twitter-auth-type,x-twitter-client-language")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Origin", "https://twitter.com")
	req.Header.Set("Referer", fmt.Sprintf("https://twitter.com/%s/followers", task.ScreenName))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36")

	setCookiesForRequest(req, task.cookies)

	return req, nil
}

func (task DownloadFollowersTask) createShowRequest() (*http.Request, error) {
	req, err := http.NewRequest("GET","https://api.twitter.com/1.1/users/show.json", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()

	q.Add("include_profile_interstitial_type", "1")
	q.Add("include_blocking", "1")
	q.Add("include_blocked_by", "1")
	q.Add("include_followed_by", "1")
	q.Add("include_want_retweets", "1")
	q.Add("include_mute_edge", "1")
	q.Add("include_can_dm", "1")
	q.Add("include_can_media_tag", "1")
	q.Add("skip_status", "1")
	q.Add("screen_name", strings.ToLower(task.ScreenName))

	req.URL.RawQuery = q.Encode()

	req.Header.Set("authorization", task.headers["authorization"])
	req.Header.Set("Origin", "https://twitter.com")
	req.Header.Set("Referer", fmt.Sprintf("https://twitter.com/%s/followers", task.ScreenName))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36")
	req.Header.Set("x-csrf-token", task.cookies["ct0"])
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-client-language", "en")

	setCookiesForRequest(req, task.cookies)

	return req, nil
}


func (task DownloadFollowersTask) createUsersOptionsRequest(userId int64, cursor string) (*http.Request, error) {
	req, err := http.NewRequest("OPTIONS","https://api.twitter.com/1.1/followers/list.json", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()

	q.Add("include_profile_interstitial_type", "1")
	q.Add("include_blocking", "1")
	q.Add("include_blocked_by", "1")
	q.Add("include_followed_by", "1")
	q.Add("include_want_retweets", "1")
	q.Add("include_mute_edge", "1")
	q.Add("include_can_dm", "1")
	q.Add("include_can_media_tag", "1")
	q.Add("skip_status", "1")
	q.Add("cursor", cursor)
	q.Add("user_id", strconv.FormatInt(userId, 10))
	q.Add("count", "20")

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Access-Control-Request-Headers", "authorization,x-csrf-token,x-twitter-active-user,x-twitter-auth-type,x-twitter-client-language")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Origin", "https://twitter.com")
	req.Header.Set("Referer", fmt.Sprintf("https://twitter.com/%s/followers", task.ScreenName))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36")

	setCookiesForRequest(req, task.cookies)

	return req, nil
}

func (task DownloadFollowersTask) createUsersRequest(userId int64, cursor string) (*http.Request, error) {
	req, err := http.NewRequest("GET", "https://api.twitter.com/1.1/followers/list.json", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("include_profile_interstitial_type", "1")
	q.Add("include_blocking", "1")
	q.Add("include_blocked_by", "1")
	q.Add("include_followed_by", "1")
	q.Add("include_want_retweets", "1")
	q.Add("include_mute_edge", "1")
	q.Add("include_can_dm", "1")
	q.Add("include_can_media_tag", "1")
	q.Add("skip_status", "1")
	q.Add("cursor", cursor)
	q.Add("user_id", strconv.FormatInt(userId, 10))
	q.Add("count", "20")

	req.URL.RawQuery = q.Encode()

	req.Header.Set("authorization", task.headers["authorization"])
	req.Header.Set("Origin", "https://twitter.com")
	req.Header.Set("Referer", fmt.Sprintf("https://twitter.com/%s/followers", task.ScreenName))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36")
	req.Header.Set("x-csrf-token", task.cookies["ct0"])
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-client-language", "en")

	setCookiesForRequest(req, task.cookies)

	return req, nil
}

