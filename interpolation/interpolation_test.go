package interpolation_test

import (
	"finala/interpolation"
	"reflect"
	"testing"
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
