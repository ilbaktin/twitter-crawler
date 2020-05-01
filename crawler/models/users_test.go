package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMinInt(t *testing.T) {
	cases := []struct {
		arr         []int
		expectedMin int
	}{
		{
			arr:         []int{0, 1, 2, 3, -10, 15, -100, 100},
			expectedMin: -100,
		},
		{
			arr:         []int{2, 4},
			expectedMin: 2,
		},
		{
			arr:         []int{},
			expectedMin: 0,
		},
	}
	for _, testCase := range cases {
		min := minInt(testCase.arr...)
		assert.Equal(t, testCase.expectedMin, min, "wrong minimal value")
	}
}

func TestBatchUsersArray(t *testing.T) {
	cases := []struct {
		arr               []*User
		batchSize         int
		expectedBatchLens []int
	}{
		{
			arr:               make([]*User, 5),
			batchSize:         2,
			expectedBatchLens: []int{2, 2, 1},
		},
		{
			arr:               make([]*User, 10),
			batchSize:         20,
			expectedBatchLens: []int{10},
		},
	}

	for _, testCase := range cases {
		batched := BatchUsersArray(testCase.arr, testCase.batchSize)
		assert.Equal(t, len(testCase.expectedBatchLens), len(batched), "wrong amount of batches")

		for i := 0; i < len(batched); i++ {
			assert.Equal(t, testCase.expectedBatchLens[i], len(batched[i]), "batch %d is wrong len", i)
		}
	}
}
