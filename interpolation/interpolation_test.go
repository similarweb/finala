package interpolation_test

import (
	"finala/interpolation"
	"fmt"
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

func TestExtractTimestamp(t *testing.T) {
	const index_prefix, timestamp = "general", 1595510218
	index_name := fmt.Sprintf("%s_%d", index_prefix, timestamp)
	extractedTimestamp, err := interpolation.ExtractTimestamp(index_name)
	if err != nil {
		t.Fatalf("error occured while running extractTimestamp, e: %s\n", err)
	}

	if extractedTimestamp != timestamp {
		t.Errorf("extractedTimestamp %d is not equal to expected timestamp %d", extractedTimestamp, timestamp)
	}
}

func TestExtractExecutionName(t *testing.T) {
	const index_prefix, timestamp = "general", 1595510218
	index_name := fmt.Sprintf("%s_%d", index_prefix, timestamp)
	extractedExecutionName, err := interpolation.ExtractExecutionName(index_name)
	if err != nil {
		t.Fatalf("error occured while running extractExecutionName, e: %s\n", err)
	}

	if extractedExecutionName != index_prefix {
		t.Errorf("extractedExecutionName %s is not equal to expected timestamp %s", extractedExecutionName, index_prefix)
	}
}

func TestExtractAccountInformation(t *testing.T) {
	const name, id = "Test", "1234567890"
	//test for right input
	accountInfo := fmt.Sprintf("%s_%s", name, id)
	extractedName, extractedId, err := interpolation.ExtractAccountInformation(accountInfo)
	if err != nil {
		t.Fatalf("error occured while running ExtractAccountInformation e: %s\n", err)
	}

	if extractedName != name {
		t.Errorf("extractedName %s is not equal to expected name %s", extractedName, name)
	}

	if extractedId != id {
		t.Errorf("extractedId %s is not equal to expected id %s", extractedId, id)
	}

	//test for wrong input
	const wrong = "noUnderScore"
	_, _, err = interpolation.ExtractAccountInformation(wrong)
	if err == nil {
		t.Errorf("ExtractAccountInformation returns no error for input without underscore: %s", wrong)
	}
}
