package crawler_tasks

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/scarecrow6977/twitter-crawler/crawler/conf"
	"github.com/scarecrow6977/twitter-crawler/crawler/log"
	"github.com/scarecrow6977/twitter-crawler/crawler/models"
	"github.com/scarecrow6977/twitter-crawler/crawler/storage"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type DownloadFollowersTask struct {
	ScreenName string

	cookies map[string]string
	headers map[string]string
	*log.Logger
	httpClient *http.Client
}

func (task *DownloadFollowersTask) HttpClient() *http.Client {
	return task.httpClient
}

func (task *DownloadFollowersTask) initTor() error {
	var torProxy string = "socks5://127.0.0.1:9050" // 9150 w/ Tor Browser
	torProxyUrl, err := url.Parse(torProxy)
	if err != nil {
		return errors.Wrapf(err, "Error parsing Tor proxy URL: %s", torProxy)
	}

	// Set up a custom HTTP transport to use the proxy and create the client
	torTransport := &http.Transport{Proxy: http.ProxyURL(torProxyUrl)}
	task.httpClient = &http.Client{
		Transport: torTransport,
		Timeout:   time.Second * 10,
	}

	return nil
}

func (task *DownloadFollowersTask) Exec(stor storage.Storage) error {
	task.Logger = log.NewLogger(fmt.Sprintf("DownloadFollowersTask '%s'", task.ScreenName))
	err := task.initTor()
	if err != nil {
		return errors.Wrapf(err, "can't init tor")
	}
	task.LogInfo("Tor initialized!")

	task.Logger.LogInfo("start downloading followers")

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
		//err = task.doShowOptionsRequest()
		//if err != nil {
		//	return nil
		//}
		user, err = task.doShowRequest()
		if err != nil {
			return err
		}
		err = stor.AddNewUsers([]*models.User{user})
		if err != nil {
			return err
		}
	} else {
		task.LogInfo("User '%s' found in db, next cursor = %s", task.ScreenName, user.NextCursorStr)
	}
	cursor := user.NextCursorStr

	for cursor != "0" {
		//err = task.doOptionsUsersRequest(user.Id, cursor)
		//if err != nil {
		//	return nil
		//}
		usersJsonResp, err := task.doUsersRequest(user.Id, cursor)
		switch err {
		case ErrLimitReached:
			sleepDur := time.Duration(config.ApiLimitTimeout) * time.Second
			time.Sleep(sleepDur)
			log.LogInfo("Sleep for %d seconds cause 429 status received", int(sleepDur.Seconds()))
			return err
		case ErrPrivateProfile:
			err = stor.UpdateUserState(user)
			log.LogInfo("Exit cause user has got private profile")
			return err
		default:
			if err != nil {
				sleepDur := 15 * time.Second
				time.Sleep(sleepDur)
				log.LogInfo("Sleep for %d seconds cause unknown error", int(sleepDur.Seconds()))
				return err
			}
		}

		users := usersJsonResp.Users
		followers := make([]*models.Follower, 0, len(users))
		for _, follower := range users {
			followers = append(followers, &models.Follower{
				UserId:     user.Id,
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
		user.AreFollowersDownloaded = usersJsonResp.NextCursor == 0
		err = stor.UpdateUserState(user)
		if err != nil {
			return err
		}

		cursor = user.NextCursorStr

		task.LogInfo("downloaded %d followers", len(followers))
	}
	task.LogInfo("Followers successfully downloaded.")

	return nil
}

func (task *DownloadFollowersTask) doShowRequest() (user *models.User, err error) {
	showReq, err := task.createShowRequest()
	if err != nil {
		return nil, err
	}
	resp, err := task.HttpClient().Do(showReq)
	if err != nil {
		return nil, fmt.Errorf("request failed, err='%v'", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 429:
		return nil, ErrLimitReached
	case 401:
		return nil, ErrPrivateProfile
	default:
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("bad response status code, got %d (%s), want 200", resp.StatusCode, resp.Status)
		}
	}

	jsonBytes, err := ioutil.ReadAll(resp.Body)

	user = &models.User{}
	err = json.Unmarshal(jsonBytes, user)
	if err != nil {
		return nil, err
	}
	//user.AdditionalData = string(jsonBytes)

	//user.AdditionalData = string(jsonBytes)
	user.NextCursor = -1
	user.NextCursorStr = "-1"

	return user, nil
}

func (task *DownloadFollowersTask) doOptionsUsersRequest(userId int64, cursor string) error {
	optionsReq, err := task.createUsersOptionsRequest(userId, cursor)
	if err != nil {
		return err
	}
	optionsResp, err := task.HttpClient().Do(optionsReq)
	if err != nil {
		return errors.Wrap(err, "options users request failed")
	}
	defer optionsResp.Body.Close()

	return nil
}

func (task *DownloadFollowersTask) doShowOptionsRequest() error {
	optionsReq, err := task.createShowOptionsRequest()
	if err != nil {
		return err
	}
	optionsResp, err := task.HttpClient().Do(optionsReq)
	if err != nil {
		return errors.Wrap(err, "options show request failed")
	}
	defer optionsResp.Body.Close()

	return nil
}

var ErrLimitReached = errors.New("api limit reached")
var ErrPrivateProfile = errors.New("user has got private profile")

func (task *DownloadFollowersTask) doUsersRequest(userId int64, cursor string) (userJsonResp *usersJsonResponse, err error) {
	usersReq, err := task.createUsersRequest(userId, cursor)
	if err != nil {
		return nil, err
	}
	resp, err := task.HttpClient().Do(usersReq)
	if err != nil {
		return nil, fmt.Errorf("request failed, err='%v'", err)
	}
	defer resp.Body.Close()
	//reqBytes, _ := httputil.DumpRequest(usersReq, true)
	//task.LogInfo("REQ:\n %v", string(reqBytes))
	//
	//respBytes, _ := httputil.DumpResponse(resp, true)
	//
	//task.LogInfo("RESP %d:\n %v", resp.StatusCode, string(respBytes))
	//reader, _ := gzip.NewReader(resp.Body)
	//
	//io.Copy(os.Stdout, reader)
	switch resp.StatusCode {
	case 429:
		return nil, ErrLimitReached
	case 401:
		return nil, ErrPrivateProfile
	default:
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("bad response status code, got %d (%s), want 200", resp.StatusCode, resp.Status)
		}
	}

	jsonBytes, err := ioutil.ReadAll(resp.Body)

	usersJsonResp := &usersJsonResponse{}
	err = json.Unmarshal(jsonBytes, usersJsonResp)
	if err != nil {
		return nil, errors.Wrap(err, "parse response users req")
	}

	return usersJsonResp, nil

}

func (task *DownloadFollowersTask) createShowOptionsRequest() (*http.Request, error) {
	req, err := http.NewRequest("OPTIONS", "https://api.twitter.com/1.1/users/show.json", nil)
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

func (task *DownloadFollowersTask) createShowRequest() (*http.Request, error) {
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

func (task *DownloadFollowersTask) createUsersOptionsRequest(userId int64, cursor string) (*http.Request, error) {
	req, err := http.NewRequest("OPTIONS", "https://api.twitter.com/1.1/followers/list.json", nil)
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
	q.Add("count", "200")

	req.URL.RawQuery = q.Encode()

	req.Header.Set("Access-Control-Request-Headers", "authorization,x-csrf-token,x-twitter-active-user,x-twitter-auth-type,x-twitter-client-language")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Origin", "https://twitter.com")
	req.Header.Set("Referer", fmt.Sprintf("https://twitter.com/%s/followers", task.ScreenName))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36")

	setCookiesForRequest(req, task.cookies)

	return req, nil
}

func (task *DownloadFollowersTask) createUsersRequest(userId int64, cursor string) (*http.Request, error) {
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
	q.Add("count", "200")

	req.URL.RawQuery = q.Encode()

	req.Header.Set("authorization", task.headers["authorization"])
	//req.Header.Set("Accept", "api.twitter.com")
	//req.Header.Set("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	req.Header.Set("Host", "api.twitter.com")
	req.Header.Set("Origin", "https://twitter.com")
	req.Header.Set("Referer", fmt.Sprintf("https://twitter.com/%s/followers", task.ScreenName))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:70.0) Gecko/20100101 Firefox/70.0")
	req.Header.Set("x-csrf-token", task.cookies["ct0"])
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-client-language", "en")
	req.Header.Set("x-twitter-polling", "true")

	setCookiesForRequest(req, task.cookies)

	return req, nil
}
