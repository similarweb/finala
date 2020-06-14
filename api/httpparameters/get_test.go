package httpparameters_test

import (
	"finala/api/httpparameters"
	"fmt"
	"net/http"
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
