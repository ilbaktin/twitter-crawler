package main

import (
	"github.com/scarecrow6977/twitter-crawler/crawler/conf"
	"github.com/scarecrow6977/twitter-crawler/crawler/graph"
	"github.com/scarecrow6977/twitter-crawler/crawler/models"
	pg_storage "github.com/scarecrow6977/twitter-crawler/crawler/storage/pg-storage"
	"log"
)

const batchSize = 5000

const start = 0
const skip = 100

func main() {
	conf.Init("config.yaml")
	config, err := conf.LoadConfig()
	if err != nil {
		log.Fatalf("Can't load config, err=%v", err)
	}

	imp, err := graph.NewNeo4jImporter(config.Neo4jAccess)
	if err != nil {
		log.Fatalf("Can't connect to neo4j db, err=%v", err)
	}
	log.Println("Connect to neo4j OK")

	pgStor, err := pg_storage.NewPgStorage(config.PostgresAccess)
	if err != nil {
		log.Fatalf("Can't connect to postgres, err=%v", err)
	}
	log.Println("Connect to postgres OK")

	//readyUsers, err := pgStor.GetUsersWithDownloadedFollowersSorted(amount, start)
	userIds, err := graph.ProcessUsersFromCsv("spark/data/subgraph_1197663692621590529_10000000.csv", 6_000_000)
	if err != nil {
		//log.Fatalf("Can't receive ready users from postgres db, err=%v", err)
		log.Fatalf("Can't receive ready users from csv, err=%v", err)
	}
	log.Printf("Users received, count=%d\n", len(userIds))

	log.Println("Get users from csv OK")
	for idx, userId := range userIds {
		if idx < skip {
			log.Printf("Skipping %d user (userId=%d)", idx, userId)
			continue
		}
		user, err := pgStor.GetUserById(userId)
		if err != nil {
			log.Fatalf("get user from db, userId=%d, err=%v", userId, err)
		}
		followers, err := pgStor.GetFollowers(user.Id)
		if err != nil {
			log.Fatalf("get followers from db, userId=%d, err=%v", user.Id, err)
		}
		log.Printf(
			"%d (count=%d): User %s(id=%d) and his followers (amount=%d) are being imported now...",
			idx+start,
			idx,
			user.ScreenName,
			user.Id,
			len(followers),
		)
		//log.Printf("received followers for user %s(id=%d)", user.ScreenName, user.Id)
		followersBatched := models.BatchUsersArray(followers, batchSize)
		importedCount := 0
		for _, followersBatch := range followersBatched {
			err = imp.ImportFollowers(user, followersBatch)
			if err != nil {
				log.Fatalf("Can't import users to neo4j, err=%v", err)
			}
			importedCount += len(followersBatch)
			log.Printf("%d of %d imported", importedCount, len(followers))
		}
		log.Println("done")

	}
	log.Println("Users with their followers successfully imported to neo4j")
}
