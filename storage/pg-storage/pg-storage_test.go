package pg_storage

import (
	"encoding/json"
	"testing"
	"time"
	"univer/twitter-crawler/conf"
	"univer/twitter-crawler/models"
)

func TestPgStorage_AddNewFollowers(t *testing.T) {
	followers := []*models.Follower {
		{1,3,},
		{1,2,},
		{1,4,},
		{2,1,},
		{3,3,},
		{4,4,},
	}
	conf.Init("../../config.yaml")
	config, err := conf.LoadConfig()
	if err != nil {
		t.Errorf("can't load config, err='%v'", err)
		return
	}
	s, err := NewPgStorage(config.PostgresAccess)
	if err != nil {
		t.Errorf("can't create pg storage, err='%v'", err)
		return
	}
	err = s.AddNewFollowers(followers)
	if err != nil {
		t.Errorf("can't save followers, err='%v'", err)
		return
	}
}

func TestPgStorage_AddNewUsers(t *testing.T) {
	data, _ := json.Marshal(map[string]interface{}{"description":1})
	users := []*models.User{
		{Id: 1, IdStr: "1", ScreenName: "first", Name: "asdasd", DateLastChange: time.Now(), AdditionalData: string(data)},
		{Id: 2, IdStr: "2", ScreenName: "second", Name: "efefe", DateLastChange: time.Now(), AdditionalData: string(data)},
		{Id: 3, IdStr: "3", ScreenName: "third", Name: "sdfsadfas", DateLastChange: time.Now(), AdditionalData: string(data)},

	}
	conf.Init("../../config.yaml")
	config, err := conf.LoadConfig()
	if err != nil {
		t.Errorf("can't load config, err='%v'", err)
		return
	}
	s, err := NewPgStorage(config.PostgresAccess)
	if err != nil {
		t.Errorf("can't create pg storage, err='%v'", err)
		return
	}
	err = s.AddNewUsers(users)
	if err != nil {
		t.Errorf("can't save users, err='%v'", err)
		return
	}
}

func TestPgStorage_GetUserById(t *testing.T) {
	conf.Init("../../config.yaml")
	config, err := conf.LoadConfig()
	if err != nil {
		t.Errorf("can't load config, err='%v'", err)
		return
	}
	s, err := NewPgStorage(config.PostgresAccess)
	if err != nil {
		t.Errorf("can't create pg storage, err='%v'", err)
		return
	}
	_, err = s.GetUserById(1)
	if err != nil {
		t.Errorf("can't get user with id=1, err='%v'", err)
	}
	_, err = s.GetUserById(4)
	if err == nil {
		t.Errorf("user with id=4 shouldn't exist, err='%v'", err)
	}
}

func TestPgStorage_GetUserByScreenName(t *testing.T) {
	conf.Init("../../config.yaml")
	config, err := conf.LoadConfig()
	if err != nil {
		t.Errorf("can't load config, err='%v'", err)
		return
	}
	s, err := NewPgStorage(config.PostgresAccess)
	if err != nil {
		t.Errorf("can't create pg storage, err='%v'", err)
		return
	}
	_, err = s.GetUserByScreenName("first")
	if err != nil {
		t.Errorf("can't get user with screen_name='first', err='%v'", err)
	}
	_, err = s.GetUserByScreenName("not_exists")
	if err == nil {
		t.Errorf("user with screen_name='not_exists' shouldn't exist, err='%v'", err)
	}
}

func TestPgStorage_UpdateUserState(t *testing.T) {
	user := &models.User{Id: 1, IdStr: "1", ScreenName: "first", Name: "asdasd"}
	user.NextCursor = 100000
	user.NextCursorStr = "100000"
	user.IsFollowersDownloaded = true
	user.DateLastChange = time.Now()

	conf.Init("../../config.yaml")
	config, err := conf.LoadConfig()
	if err != nil {
		t.Errorf("can't load config, err='%v'", err)
		return
	}
	s, err := NewPgStorage(config.PostgresAccess)
	if err != nil {
		t.Errorf("can't create pg storage, err='%v'", err)
		return
	}
	err = s.UpdateUserState(user)
	if err != nil {
		t.Errorf("can't update user state, err='%v'", err)
		return
	}
}