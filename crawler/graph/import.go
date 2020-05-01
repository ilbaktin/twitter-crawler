package graph

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/pkg/errors"
	"github.com/scarecrow6977/twitter-crawler/crawler/conf"
	"github.com/scarecrow6977/twitter-crawler/crawler/log"
	"github.com/scarecrow6977/twitter-crawler/crawler/models"
)

type Neo4jImporter struct {
	driver  neo4j.Driver
	session neo4j.Session
	*log.Logger
}

func NewNeo4jImporter(config conf.Neo4jAccessConfig) (*Neo4jImporter, error) {
	driver, err := neo4j.NewDriver(config.Uri, neo4j.BasicAuth(config.User, config.Password, ""))
	if err != nil {
		return nil, err
	}

	session, err := driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, err
	}

	imp := &Neo4jImporter{
		driver:  driver,
		session: session,
	}
	imp.Logger = log.NewLogger("Neo4j Importer")
	fmt.Println(config.User, config.Password, config.Uri)
	return imp, nil
}

func (importer *Neo4jImporter) Close() error {
	err := importer.session.Close()
	if err != nil {
		return err
	}
	return importer.driver.Close()
}

func (importer *Neo4jImporter) ImportUsers(users []*models.User) error {
	_, err := importer.session.WriteTransaction(func(tx neo4j.Transaction) (i interface{}, e error) {
		for _, user := range users {
			_, err := tx.Run(
				`MERGE (u:User {id: $id}) 
						SET u.screen_name = $screen_name, u.location=$location, 
						u.followers_count=$followers_count`,
				map[string]interface{}{
					"screen_name":     user.ScreenName,
					"location":        user.Location,
					"id":              user.Id,
					"followers_count": user.FollowersCount,
				},
			)
			if err != nil {
				return nil, err
			}

			fmt.Printf("Imported user: screen_name=%s, err=%v\n", user.ScreenName, err)
		}

		return nil, nil
	})
	return err
}

func (importer *Neo4jImporter) ImportFollowers(user *models.User, followers []*models.User) error {
	_, err := importer.session.WriteTransaction(func(tx neo4j.Transaction) (i interface{}, e error) {
		err := addUserTx(tx, user)
		if err != nil {
			return nil, errors.Wrapf(err, "add user to graph, userId=%d", user.Id)
		}
		for _, follower := range followers {
			err := addUserTx(tx, follower)
			if err != nil {
				return nil, errors.Wrapf(err, "add user to graph, userId=%d", user.Id)
			}
			err = addFollowerTx(tx, user.Id, follower.Id)
			if err != nil {
				return nil, errors.Wrapf(err, "add follower of user to graph, userId=%d, followerId=%d", user.Id, follower.Id)
			}
		}

		return nil, nil
	})
	return err
}

func addUserTx(tx neo4j.Transaction, user *models.User) error {
	_, err := tx.Run(
		`MERGE (u:User { id: $id }) 
						SET u.screen_name = $screen_name, u.location=$location, 
						u.followers_count=$followers_count`,
		map[string]interface{}{
			"screen_name":     user.ScreenName,
			"location":        user.Location,
			"id":              user.Id,
			"followers_count": user.FollowersCount,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func addFollowerTx(tx neo4j.Transaction, userId, followerId int64) error {
	_, err := tx.Run(
		`MATCH (user:User { id: $user_id }),(follower:User { id: $follower_id })
				MERGE (follower)-[r:FOLLOWS]->(user)`,
		map[string]interface{}{
			"user_id":     userId,
			"follower_id": followerId,
		},
	)
	if err != nil {
		return err
	}
	return nil
}
