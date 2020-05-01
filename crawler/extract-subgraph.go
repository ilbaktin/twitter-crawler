package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/scarecrow6977/twitter-crawler/crawler/conf"
	"github.com/scarecrow6977/twitter-crawler/crawler/log"
	pg_storage "github.com/scarecrow6977/twitter-crawler/crawler/storage/pg-storage"
	"os"
)

type follow struct {
	userId     int64
	followerId int64
}

func extractSubgraph(pgStorage *pg_storage.PgStorage, firstUserId, size int64) error {
	doneCh := make(chan struct{})
	filename := fmt.Sprintf("subgraph_%d_%d.csv", firstUserId, size)
	followChan := getFollowsChan(pgStorage, doneCh, firstUserId, size)
	err := consumeFollows(followChan, doneCh, filename)
	if err != nil {
		return errors.Wrap(err, "consume follows")
	}
	return nil
}

func getFollowsChan(pgStorage *pg_storage.PgStorage, doneCh chan<- struct{}, firstUserId, size int64) (follows <-chan follow) {
	userIdChan := make(chan int64, 2*size)
	followChan := make(chan follow, 1000)
	userIdChan <- firstUserId

	go func() {
		var edgesCount int64 = 0
		visited := make(map[int64]struct{})
	Top:
		for {
			select {
			case userId := <-userIdChan:
				followerIds, err := pgStorage.GetFollowerIds(userId)
				if err != nil {
					log.LogError("get follower ids of user %d, err=%v", userId, err)
					continue
				}
				//log.LogInfo("Received followers of user %v (amount %v), edgesCount=%v", userId, len(followerIds), edgesCount)
				visited[userId] = struct{}{}
				for _, followerId := range followerIds {
					if _, ok := visited[followerId]; !ok {
						userIdChan <- followerId
					}
					followChan <- follow{
						userId:     userId,
						followerId: followerId,
					}
					edgesCount++
				}
				if edgesCount > size {
					close(userIdChan)
					close(followChan)
					doneCh <- struct{}{}
					break Top
				}
			}
		}
	}()

	return followChan
}

func consumeFollows(follows <-chan follow, doneCh <-chan struct{}, filename string) error {
	csvFile, err := os.Create(filename)

	if err != nil {
		return errors.Wrapf(err, "create file %s", filename)
	}

	writer := csv.NewWriter(csvFile)
	counter := 0
	for {
		select {
		case f := <-follows:
			data := []string{
				fmt.Sprintf("%v", f.userId),
				fmt.Sprintf("%v", f.followerId),
			}
			err := writer.Write(data)
			if err != nil {
				return err
			}
			counter++
			if counter%1000 == 0 {
				writer.Flush()
			}
		case <-doneCh:
			writer.Flush()

			return nil
		}
	}
}

func main() {
	var configPath string
	var userId int64
	var size int64
	flag.StringVar(&configPath, "config", "config.yaml", "path to the config file")
	flag.Int64Var(&userId, "from", 0, "id of node search starts from")
	flag.Int64Var(&size, "size", 1000, "size of resulting subgraph")

	flag.Parse()

	conf.Init(configPath)
	config, err := conf.LoadConfig()
	if err != nil {
		log.LogError("can't load config, err='%v'", err)
		return
	}
	log.SetVerbosityLevel(2)

	pgStorage, err := pg_storage.NewPgStorage(config.PostgresAccess)
	if err != nil {
		log.LogError(fmt.Sprintf("can't load config, err='%v'", err))
		return
	}

	err = extractSubgraph(pgStorage, userId, size)
	if err != nil {
		log.LogError("can't extract subgraph, err=%v", err)
	}
}
