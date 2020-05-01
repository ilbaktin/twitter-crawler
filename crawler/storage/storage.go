package storage

import (
	"github.com/scarecrow6977/twitter-crawler/crawler/models"
	"github.com/scarecrow6977/twitter-crawler/crawler/storage/fs-storage"
)

type StorageOld interface {
	Save(record *fs_storage.Record) error
	GetLastCursor() (string, error)
	SetLastCursor(cursor string) error
}

type Storage interface {
	AddNewFollowers(followers []*models.Follower) error
	AddNewUsers(users []*models.User) error
	UpdateUserState(user *models.User) error
	GetUserById(id int64) (*models.User, error)
	GetUserByScreenName(screenName string) (*models.User, error)
	GetUsersWithNotDownloadedFollowers(n int64) ([]*models.User, error)
	GetUsersWithDownloadedFollowers(n int64) ([]*models.User, error)
	GetUsersWithNotDownloadedFollowersSorted(n, offset int64) ([]*models.User, error)
	GetUsersWithDownloadedFollowersSorted(n, offset int64) ([]*models.User, error)
}
