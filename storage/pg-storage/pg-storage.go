package pg_storage

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"net/url"
	"time"
	"univer/twitter-crawler/conf"
	"univer/twitter-crawler/models"
)

type PgStorage struct {
	pgConn *sqlx.DB
}

func NewPgStorage(config conf.PostgresAccessConfig) (*PgStorage, error) {
	var userInfo *url.Userinfo
	if config.Password != nil {
		userInfo = url.UserPassword(config.User, *config.Password)
	} else {
		userInfo = url.User(config.User)
	}
	host := config.Host
	if config.Port != nil {
		host = fmt.Sprintf("%s:%d", host, *config.Port)
	}
	connUrl := url.URL{
		Scheme: "postgres",
		Host:   host,
		User:   userInfo,
		Path:   config.Dbname,
	}
	q := connUrl.Query()
	q.Add("sslmode", "disable")
	connUrl.RawQuery = q.Encode()
	pgConn, err := sqlx.Open("postgres", connUrl.String())
	if err != nil {
		return nil, err
	}
	return &PgStorage{
		pgConn: pgConn,
	}, nil
}

func (s *PgStorage) SaveFollower(userId, followerId int64) error {
	follower := &models.Follower{
		UserId:     userId,
		FollowerId: followerId,
	}
	_, err := s.pgConn.NamedExec("INSERT INTO followers (user_id, follower_id) VALUES (:user_id, :follower_id)", follower)
	if err != nil {
		return err
	}
	return nil
}

func (s *PgStorage) AddNewFollowers(followers []*models.Follower) error {
	tx, err := s.pgConn.Beginx()
	if err != nil {
		return err
	}
	var txErr error
	defer func() {
		if txErr != nil {
			tx.Rollback()
		}
	}()
	stmt, txErr := tx.Preparex(`INSERT INTO followers (user_id, follower_id) VALUES ($1, $2) ON CONFLICT ON CONSTRAINT connection DO NOTHING`) // pq.CopyIn("followers", "user_id", "follower_id")
	if txErr != nil {
		return txErr
	}
	for _, follower := range followers {
		_, txErr = stmt.Exec(follower.UserId, follower.FollowerId)
		if txErr != nil {
			return txErr
		}
	}
	_, txErr = stmt.Exec()
	if err != nil {
		return txErr
	}
	txErr = stmt.Close()
	if err != nil {
		return txErr
	}
	txErr = tx.Commit()
	return txErr
}

func (s *PgStorage) AddNewUsers(users []*models.User) error {
	tx, err := s.pgConn.Beginx()
	if err != nil {
		return err
	}
	var txErr error
	defer func() {
		if txErr != nil {
			tx.Rollback()
		}
	}()
	stmt, txErr := tx.PrepareNamed(
		`
INSERT INTO users (id, id_str, screen_name, name, created_at, followers_count, friends_count, verified, date_last_change, protected, location) 
VALUES (:id, :id_str, :screen_name, :name, :created_at, :followers_count, :friends_count, :verified, :date_last_change, :protected, :location) 
ON CONFLICT ON CONSTRAINT users_pkey DO NOTHING`)
	if txErr != nil {
		return txErr
	}
	for _, user := range users {
		user.DateLastChange = time.Now()
		_, txErr = stmt.Exec(user)
		if txErr != nil {
			return txErr
		}
	}
	//_, txErr = stmt.Exec()
	//if err != nil {
	//	return txErr
	//}
	txErr = stmt.Close()
	if err != nil {
		return txErr
	}
	txErr = tx.Commit()
	return txErr
}

func (s *PgStorage) UpdateUserState(user *models.User) error {
	user.DateLastChange = time.Now()
	_, err := s.pgConn.NamedExec(
		`UPDATE users SET (next_cursor, next_cursor_str, are_followers_downloaded, date_last_change, protected, location) = 
(:next_cursor, :next_cursor_str, :are_followers_downloaded, :date_last_change, :protected, :location) WHERE id=:id
`, user)
	return err
}

func (s *PgStorage) GetUserById(id int64) (*models.User, error) {
	user := &models.User{}
	err := s.pgConn.Get(user, "SELECT * FROM users WHERE users.id=$1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *PgStorage) GetUserByScreenName(screenName string) (*models.User, error) {
	user := &models.User{}
	err := s.pgConn.Get(user, "SELECT * FROM users WHERE users.screen_name=$1 LIMIT 1", screenName)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *PgStorage) GetUsersWithNotDownloadedFollowers(n int64) ([]*models.User, error) {
	users := make([]*models.User, 0, n)
	err := s.pgConn.Select(&users, "SELECT * FROM users WHERE users.are_followers_downloaded=false LIMIT $1", n)
	if err != nil {
		return nil, err
	}
	return users, nil
}
