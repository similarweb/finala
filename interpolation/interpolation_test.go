package interpolation_test

import (
	"finala/interpolation"
	"reflect"
	"testing"

	awsClient "github.com/aws/aws-sdk-go/aws"
)

func TestGetUnique(t *testing.T) {
	t.Parallel()
	stringsList := []string{"Hi", "Bye", "Hi", "No", "yes"}
	uniqueList := interpolation.UniqueStr(stringsList)
	expectedList := []string{"Hi", "Bye", "No", "yes"}

	if !reflect.DeepEqual(uniqueList, expectedList) {
		t.Errorf("uniqueStr function did not retrive a unique list, uniqueList: %s, expectedList: %s.", uniqueList, expectedList)
	}
}

func TestChunkIterator(t *testing.T) {
	stringsList := []*string{awsClient.String("Hi"),
		awsClient.String("Bye"),
		awsClient.String("Hi"),
		awsClient.String("No"),
		awsClient.String("yes"),
		awsClient.String("A"),
		awsClient.String("B"),
	}
	chunkSize := 2
	expectedResults := map[int][]*string{
		1: {awsClient.String("Hi"), awsClient.String("Bye")},
		2: {awsClient.String("Hi"), awsClient.String("No")},
		3: {awsClient.String("yes"), awsClient.String("A")},
		4: {awsClient.String("B")},
	}
	stringIterator := interpolation.ChunkIterator(stringsList, chunkSize)
	var counter int = 1
	for stringBatch := stringIterator(); stringBatch != nil; stringBatch = stringIterator() {
		if !reflect.DeepEqual(stringBatch, expectedResults[counter]) {
			t.Errorf("stringBatch unexpected list values retrived got: %v, wanted: %v.", stringBatch, expectedResults[counter])
		}
		counter++
	}
}
