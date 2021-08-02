package api

import (
	"encoding/json"
	"finala/api/httpparameters"
	"finala/api/storage"
	"fmt"
	"github.com/golang-jwt/jwt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	queryParamFilterPrefix     = "filter_"
	resourceTrendsLimitDefault = 60
)

// DetectEventsInfo descrive the incoming HTTP events
type DetectEventsInfo struct {
	ResourceName string
	EventType    string
	EventTime    int64
	Data         interface{}
}

// GetSummary return list of summary executions
func (server *Server) GetSummary(resp http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	params := mux.Vars(req)
	executionID := params["executionID"]
	filters := httpparameters.GetFilterQueryParamWithOutPrefix(queryParamFilterPrefix, queryParams)

	response, err := server.storage.GetSummary(executionID, filters)
	if err != nil {
		server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
		return

	}
	server.JSONWrite(resp, http.StatusOK, response)
}

// GetExecutions return list collector executions
func (server *Server) GetExecutions(resp http.ResponseWriter, req *http.Request) {
	querylimit, _ := strconv.Atoi(httpparameters.QueryParamWithDefault(req, "querylimit", storage.GetExecutionsQueryLimit))
	results, err := server.storage.GetExecutions(querylimit)
	if err != nil {
		server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
		return

	}
	server.JSONWrite(resp, http.StatusOK, results)
}

func (server *Server) GetAccounts(resp http.ResponseWriter, req *http.Request) {
	queryLimit, _ := strconv.Atoi(httpparameters.QueryParamWithDefault(req, "querylimit", storage.GetExecutionsQueryLimit))
	params := mux.Vars(req)
	executionID := params["executionID"]
	accounts, err := server.storage.GetAccounts(executionID, queryLimit)

	if err != nil {
		server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
		return

	}
	server.JSONWrite(resp, http.StatusOK, accounts)
}

// GetResourceData return resuts details by resource type
func (server *Server) GetResourceData(resp http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	queryErrs := url.Values{}
	params := mux.Vars(req)
	resourceType := params["type"]
	filters := httpparameters.GetFilterQueryParamWithOutPrefix(queryParamFilterPrefix, queryParams)

	executionID := req.URL.Query().Get("executionID")
	if executionID == "" {
		queryErrs.Add("executionID", "executionID field is mandatory")
	}

	if len(queryErrs) > 0 {
		server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{ErrorQuery: queryErrs})
		return
	}

	response, err := server.storage.GetResources(resourceType, executionID, filters)
	if err != nil {
		server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
		return

	}
	server.JSONWrite(resp, http.StatusOK, response)
}

// GetResourceTrends return trends by resource type, id, region and metric
func (server *Server) GetResourceTrends(resp http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	params := mux.Vars(req)
	resourceType := params["type"]
	filters := httpparameters.GetFilterQueryParamWithOutPrefix(queryParamFilterPrefix, queryParams)

	limitString := req.URL.Query().Get("limit")
	var limit int = resourceTrendsLimitDefault
	var err error
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil || limit < 1 {
			limit = resourceTrendsLimitDefault
		}
	}

	trends, err := server.storage.GetResourceTrends(resourceType, filters, limit)
	if err != nil {
		server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
		return

	}
	server.JSONWrite(resp, http.StatusOK, trends)
}

// GetExecutionTags return resuts details by resource type
func (server *Server) GetExecutionTags(resp http.ResponseWriter, req *http.Request) {

	params := mux.Vars(req)
	executionID := params["executionID"]

	response, err := server.storage.GetExecutionTags(executionID)
	if err != nil {
		server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
		return

	}
	server.JSONWrite(resp, http.StatusOK, response)
}

// DetectEvents save collectors events data
func (server *Server) DetectEvents(resp http.ResponseWriter, req *http.Request) {

	params := mux.Vars(req)
	executionID := params["executionID"]

	buf, bodyErr := ioutil.ReadAll(req.Body)

	if bodyErr != nil {
		server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{Error: bodyErr.Error()})
		return
	}

	var detectEventsInfo []DetectEventsInfo
	err := json.Unmarshal(buf, &detectEventsInfo)
	if err != nil {
		server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{Error: err.Error()})
		return
	}

	log.WithFields(log.Fields{
		"events": len(detectEventsInfo),
	}).Info("Got bulk events")

	go func() {
		for _, event := range detectEventsInfo {

			rowData := storage.EventRow{
				ExecutionID:  executionID,
				ResourceName: event.ResourceName,
				EventType:    event.EventType,
				EventTime:    event.EventTime,
				Timestamp:    time.Now(),
				Data:         event.Data,
			}
			bolB, _ := json.Marshal(rowData)
			server.storage.Save(string(bolB))
		}
	}()

	server.JSONWrite(resp, http.StatusAccepted, nil)

}

func (server *Server) Login(resp http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		server.JSONWrite(resp, http.StatusOK, nil)
		return
	}
	if server.authentication.Enabled {
		buf, bodyErr := ioutil.ReadAll(req.Body)

		if bodyErr != nil {
			server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{Error: bodyErr.Error()})
			return
		}

		var detectUser map[string]string
		err := json.Unmarshal(buf, &detectUser)
		if err != nil {
			server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{Error: err.Error()})
			return
		}

		for _, user := range server.authentication.Accounts {
			if detectUser["Username"] == user.Name && detectUser["Password"] == user.Password {

				expTime := time.Now().Add(time.Hour * 1)

				atClaims := jwt.MapClaims{}
				atClaims["authorized"] = true
				atClaims["user_id"] = user.Name
				atClaims["exp"] = expTime.Unix()
				at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
				token, err := at.SignedString([]byte("secret"))
				if err != nil {
					server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
					return
				}

				cookie := http.Cookie{
					Name:     "jwt",
					Value:    token,
					Expires:  expTime,
					SameSite: http.SameSiteLaxMode,
				}

				http.SetCookie(resp, &cookie)
				return
			}
		}
		server.JSONWrite(resp, http.StatusUnauthorized, "{\"message\":\"Login data not authorized\"}")
	} else {
		expTime := time.Now().Add(time.Hour * 1)

		atClaims := jwt.MapClaims{}
		atClaims["authorized"] = true
		atClaims["user_id"] = "user"
		atClaims["exp"] = expTime.Unix()
		at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
		token, err := at.SignedString([]byte("secret"))
		if err != nil {
			server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
			return
		}

		cookie := http.Cookie{
			Name:     "jwt",
			Value:    token,
			Expires:  expTime,
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(resp, &cookie)
	}
}

//NotFoundRoute return when route not found
func (server *Server) NotFoundRoute(resp http.ResponseWriter, req *http.Request) {
	server.JSONWrite(resp, http.StatusNotFound, HttpErrorResponse{Error: "Path not found"})
}

//HealthCheckHandler return ok if server is up
func (server *Server) HealthCheckHandler(resp http.ResponseWriter, req *http.Request) {
	server.JSONWrite(resp, http.StatusOK, HealthResponse{Status: true})
}

// VersionHandler returns the latest Finala version
func (server *Server) VersionHandler(resp http.ResponseWriter, req *http.Request) {
	version, err := server.version.Get()
	if err != nil {
		server.JSONWrite(resp, http.StatusNotFound, HttpErrorResponse{Error: "Version was not found"})
		return
	}
	server.JSONWrite(resp, http.StatusOK, version)
}

func (server *Server) middleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("jwt")

		if err != nil {
			server.JSONWrite(resp, http.StatusUnauthorized, HttpErrorResponse{Error: "Authorize cookie not found"})
			return
		}
		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})

		if err != nil {
			server.JSONWrite(resp, http.StatusUnauthorized, HttpErrorResponse{Error: err.Error()})
			return
		}

		next.ServeHTTP(resp, req)
	})
}

//Returns json thingy wingy dingy i dont know how
func (server *Server) GetReport(resp http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	params := mux.Vars(req)
	executionID := params["executionID"]
	filters := httpparameters.GetFilterQueryParamWithOutPrefix(queryParamFilterPrefix, queryParams)

	log.WithFields(log.Fields{
		"filter": filters,
	}).Info("filter")

	filterForSummary := make(map[string]string)
	for filterKey, filterValue := range filters {
		filterForSummary[filterKey] = filterValue
	}

	response, err := server.storage.GetSummary(executionID, filterForSummary)
	if err != nil {
		server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
		return

	}

	var result []map[string]interface{}
	var attributeList []string

	for resourceName, resourceSummary := range response {
		fmt.Println(resourceName, resourceSummary) //sanity test
		if resourceSummary.ResourceCount > 0 {
			resourcesList, err := server.storage.GetResources(resourceName, executionID, filters)
			if err != nil {
				server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
				continue
			}

			log.WithFields(log.Fields{
				"name":        resourceName,
				"executionID": executionID,
				"filter":      filters,
			}).Info(resourcesList)

			//can still fail for some odd reason so we check if the resource is actually there
			if resourcesList[0] == nil {
				server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
				return
			}
			data, ok := resourcesList[0]["Data"].(map[string]interface{})
			if !ok {
				//screw your log
				continue
			}
			for key := range data {
				if key == "Tag" {
					continue
				}
				exists := false
				for index := range attributeList {
					if attributeList[index] == key {
						exists = true
					}
				}
				if !exists {
					attributeList = append(attributeList, key)
				}
			}

		}
	}

	for resourceName, resourceSummary := range response {
		fmt.Println(resourceName, resourceSummary) //sanity test
		if resourceSummary.ResourceCount > 0 {
			resourcesList, err := server.storage.GetResources(resourceName, executionID, filters)
			if err != nil {
				server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
				continue
			}
			for _, element := range resourcesList {
				data, ok := element["Data"].(map[string]interface{})
				if !ok {
					//screw your log
					continue
				}

				delete(data, "Tag")

				for _, attrName := range attributeList {
					_, ok := data[attrName]
					if !ok {
						data[attrName] = nil
					}
				}
				result = append(result, data)
			}

		}
	}

	server.JSONWrite(resp, http.StatusOK, result)
}
