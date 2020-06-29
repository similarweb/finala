package httpparameters_test

import (
	"finala/api/httpparameters"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestQueryParamWithDefault(t *testing.T) {

	expectedValue := "value"
	req, _ := http.NewRequest("GET", fmt.Sprintf("127.0.0.1?foo=%s", expectedValue), nil)

	t.Run("found", func(t *testing.T) {

		parameterValue := httpparameters.QueryParamWithDefault(req, "foo", "2")
		if parameterValue != expectedValue {
			t.Fatalf("unexpected query value parameter, got %s expected %s", parameterValue, expectedValue)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		defaultValue := "notfound"
		parameterValue := httpparameters.QueryParamWithDefault(req, "not", defaultValue)
		if parameterValue != defaultValue {
			t.Fatalf("unexpected query value parameter, got %s expected %s", parameterValue, defaultValue)
		}
	})

}

func TestGetFilterQueryParamWithOutPrefix(t *testing.T) {
	queryPrefix := "bla_"
	expectedValue := "stays"
	expectedFiltersLen := 1
	v := url.Values{}
	v.Set(fmt.Sprintf("wrongprefix_%s", expectedValue), "value")
	v.Set(fmt.Sprintf("bla_%s", expectedValue), "value")
	filters := httpparameters.GetFilterQueryParamWithOutPrefix(queryPrefix, v)

	t.Run("check_expected", func(t *testing.T) {
		if _, ok := filters[expectedValue]; !ok {
			t.Fatalf("The map of filters has unexpected filter key name, got:%v expected: %s", filters, expectedValue)
		}
		if len(filters) != expectedFiltersLen {
			t.Fatalf("The amount of filters is unexpected, got:%d expected: %d", len(filters), expectedFiltersLen)
		}
	})
}
