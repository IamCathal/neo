package app

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMapOfShortestIDs(t *testing.T) {
	IDsSlice := []int64{1, 15, 78}
	expectedMap := make(map[string]bool)
	for _, val := range IDsSlice {
		expectedMap[fmt.Sprint(val)] = true
	}

	actualMap := getMapOfShortestIDs(IDsSlice)

	assert.Equal(t, expectedMap, actualMap)
}

func TestIfIsTrue(t *testing.T) {
	steamID := 12314234535354
	myMap := make(map[string]bool)
	myMap[fmt.Sprint(steamID)] = true

	assert.True(t, ifIsTrue(fmt.Sprint(steamID), myMap))
}

func TestIndexOf(t *testing.T) {
	mySlice := []int64{56, 23, 1337, 263}
	expectedIndex := 2

	assert.Equal(t, expectedIndex, indexOf(1337, mySlice))
}

func TestToInt64(t *testing.T) {
	steamID := "123423454"
	expected := int64(123423454)

	assert.Equal(t, expected, toInt64(steamID))
}
