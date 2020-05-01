package graph

import (
	"bufio"
	"github.com/pkg/errors"
	"log"
	"os"
	"strconv"
	"strings"
)

func ProcessUsersFromCsv(filename string, preallocatedSize int) ([]int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "read file")
	}
	defer file.Close()

	userIds := make([]int64, 0, preallocatedSize)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ",")
		if len(parts) != 2 {
			continue
		}
		userId, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			log.Printf("can't parse %s to int64\n", parts[0])
			continue
		}
		userIds = append(userIds, userId)
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scanner error")
	}

	userIds = unique(userIds)

	return userIds, nil
}

func unique(intSlice []int64) []int64 {
	keys := make(map[int64]bool)
	list := []int64{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
