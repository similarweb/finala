package httpparameters

import "net/http"

// QueryParamWithDefault return query params, if not found default value will returned
func QueryParamWithDefault(req *http.Request, paramName, defaultValue string) string {
	params := req.URL.Query()
	if _, exists := params[paramName]; exists {
		return params.Get(paramName)
	}
	return defaultValue
}
