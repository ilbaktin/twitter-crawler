package crawler_tasks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"univer/twitter-crawler/conf"
	"univer/twitter-crawler/log"
	"univer/twitter-crawler/models"
	"univer/twitter-crawler/storage"
)

type DownloadTweetsTask struct {
	ScreenName string

	cookies map[string]string
	headers map[string]string
	*log.Logger
}

func (task DownloadTweetsTask) Exec(stor storage.Storage) error {
	task.Logger = log.NewLogger(fmt.Sprintf("DownloadTweetsTask '%s'", task.ScreenName))

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
		task.LogInfo("User '%s' not found in db, requesting...", task.ScreenName)
		user, err = task.doShowRequest()
		if err != nil {
			return err
		}
		err = stor.AddNewUsers([]*models.User{user})
		if err != nil {
			return err
		}
	}
	cursor := ""

	tweetJsonResp, err := task.doTweetsRequest(user.Id, cursor)
	if err != nil {
		return err
	}
	tweets := tweetJsonResp.GlobalObjects.Tweets
	users := tweetJsonResp.GlobalObjects.Users

	dataToSave := struct {
		Tweets map[string]*models.Tweet
		Users  map[string]*models.User
	}{
		Tweets: tweets,
		Users:  users,
	}
	jsonBytes, err := json.MarshalIndent(dataToSave, "", "\t")

	fileName := "data/tweets/test.json" //fmt.Sprintf("data/tweets/%s", task.ScreenName)
	err = ioutil.WriteFile(fileName, jsonBytes, os.ModePerm)
	if err != nil {
		return err
	}

	task.LogInfo("Tweets successfully downloaded for user %s.", task.ScreenName)

	return nil
}

func (task DownloadTweetsTask) doShowRequest() (user *models.User, err error) {
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

func (task DownloadTweetsTask) doTweetsRequest(userId int64, cursor string) (userJsonResp *tweetJsonResponse, err error) {
	tweetsReq, err := task.createTweetsRequest(userId, cursor)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(tweetsReq)
	if err != nil {
		return nil, fmt.Errorf("request failed, err='%v'", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad response status code, got %d, want 200", resp.StatusCode)
	}

	jsonBytes, err := ioutil.ReadAll(resp.Body)

	tweetsJsonResp := &tweetJsonResponse{}
	err = json.Unmarshal(jsonBytes, tweetsJsonResp)
	if err != nil {
		return nil, err
	}

	return tweetsJsonResp, nil

}

func (task DownloadTweetsTask) createShowRequest() (*http.Request, error) {
	req, err := http.NewRequest("GET", "https://api.twitter.com/1.1/users/show.json", nil)
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

func (task DownloadTweetsTask) createTweetsRequest(userId int64, cursor string) (*http.Request, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/2/timeline/profile/%d.json", userId), nil)
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

	q.Add("cards_platform", "Web-12")
	q.Add("include_cards", "1")
	q.Add("include_composer_source", "true")
	q.Add("include_ext_alt_text", "true")
	q.Add("include_reply_count", "1")
	q.Add("tweet_mode", "extended")
	q.Add("include_entities", "true")
	q.Add("include_user_entities", "true")
	q.Add("include_ext_media_color", "true")
	q.Add("include_ext_media_availability", "true")
	q.Add("send_error_codes", "true")
	q.Add("include_tweet_replies", "false")

	q.Add("userId", strconv.FormatInt(userId, 10))
	q.Add("count", "200")

	if cursor != "" {
		q.Add("cursor", cursor)
	}

	q.Add("ext", "mediaStats,highlightedLabel,cameraMoment")

	req.URL.RawQuery = q.Encode()

	req.Header.Set("authorization", task.headers["authorization"])
	req.Header.Set("Origin", "https://twitter.com")
	req.Header.Set("Referer", fmt.Sprintf("https://twitter.com/%s", task.ScreenName))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36")
	req.Header.Set("x-csrf-token", task.cookies["ct0"])
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-client-language", "en")

	setCookiesForRequest(req, task.cookies)

	return req, nil
}
