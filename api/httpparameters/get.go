package httpparameters

import (
	"net/http"
	"net/url"
	"strings"
)

// QueryParamWithDefault return query params, if not found default value will returned
func QueryParamWithDefault(req *http.Request, paramName, defaultValue string) string {
	params := req.URL.Query()
	if _, exists := params[paramName]; exists {
		return params.Get(paramName)
	}
	return defaultValue
}

// GetFilterQueryParamWithOutPrefix will take only the query params with a filter prefix
// And return a map of filters
func GetFilterQueryParamWithOutPrefix(queryParamFilterPrefix string, queryParams url.Values) map[string]string {
	filters := map[string]string{}
	for queryParam, value := range queryParams {
		if strings.HasPrefix(queryParam, queryParamFilterPrefix) {
			filters[strings.TrimPrefix(queryParam, queryParamFilterPrefix)] = value[0]
		}
	}
	return filters
}
