package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/neosteamfriendgraphing/common"
	"go.uber.org/zap"
)

func LogBasicInfo(msg string, req *http.Request, statusCode int) {
	vars := mux.Vars(req)
	requestStartTime, _ := strconv.ParseInt(vars["requestStartTime"], 10, 64)
	configuration.Logger.Info(msg,
		zap.String("requestID", vars["requestID"]),
		zap.Int("status", http.StatusInternalServerError),
		zap.Int64("duration", configuration.GetCurrentTimeInMs()-requestStartTime),
		zap.String("path", req.URL.EscapedPath()),
	)
}

func LogBasicErr(err error, req *http.Request, statusCode int) {
	vars := mux.Vars(req)
	requestStartTime, _ := strconv.ParseInt(vars["requestStartTime"], 10, 64)
	configuration.Logger.Error(fmt.Sprintf("%v", err),
		zap.String("requestID", vars["requestID"]),
		zap.Int("status", http.StatusInternalServerError),
		zap.Int64("duration", configuration.GetCurrentTimeInMs()-requestStartTime),
		zap.String("path", req.URL.EscapedPath()),
	)
}

func LogBasicFatal(err error, req *http.Request, statusCode int) {
	vars := mux.Vars(req)
	requestStartTime, _ := strconv.ParseInt(vars["requestStartTime"], 10, 64)
	configuration.Logger.Fatal(fmt.Sprintf("%v", err),
		zap.String("requestID", vars["requestID"]),
		zap.Int("status", http.StatusInternalServerError),
		zap.Int64("duration", configuration.GetCurrentTimeInMs()-requestStartTime),
		zap.String("path", req.URL.EscapedPath()),
	)
}

func GetUserFromDatastore(steamID string) (common.UserDocument, error) {
	targetURL := fmt.Sprintf("%s/getuser/%s", os.Getenv("DATASTORE_URL"), steamID)
	req, err := http.NewRequest("GET", targetURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", "something")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return common.UserDocument{}, err
	}
	// If no user exists in the DB (HTTP 404)
	if res.StatusCode == http.StatusNotFound {
		return common.UserDocument{}, nil
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return common.UserDocument{}, err
	}

	userDoc := datastructures.GetUserDTO{}
	err = json.Unmarshal(body, &userDoc)
	if err != nil {
		return common.UserDocument{}, err
	}

	return userDoc.User, nil
}

func JobIsNotLevelOneAndNotMax(job datastructures.Job) bool {
	return (job.MaxLevel != 1) || (job.CurrentLevel < job.MaxLevel)
}
