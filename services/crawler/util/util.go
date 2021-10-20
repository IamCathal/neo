package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/datastructures"
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

func SendBasicErrorResponse(w http.ResponseWriter, req *http.Request, err error, vars map[string]string, statusCode int) {
	w.WriteHeader(http.StatusInternalServerError)
	response := struct {
		Error string `json:"error"`
	}{
		fmt.Sprintf("Give the code monkeys this ID: '%s'", vars["requestID"]),
	}
	json.NewEncoder(w).Encode(response)
}

func SendBasicInvalidResponse(w http.ResponseWriter, req *http.Request, msg string, vars map[string]string, statusCode int) {
	w.WriteHeader(statusCode)
	response := struct {
		Error string `json:"error"`
	}{
		msg,
	}
	json.NewEncoder(w).Encode(response)
}

// GetAndRead executes a HTTP GET request and returns the body
// of the response in []byte format
func GetAndRead(URL string) ([]byte, error) {
	res, err := http.Get(URL)
	if err != nil {
		return []byte{}, err
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}

func GetFriendsFromDatastore(steamID int64) (datastructures.FriendsFromDB, error) {
	targetURL := fmt.Sprintf("%s/getfriends", os.Getenv("DATASTORE_URL"))
	reqBody := []byte(fmt.Sprintf(`{"steamID":%d}`, steamID))

	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", "something")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return datastructures.FriendsFromDB{}, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return datastructures.FriendsFromDB{}, err
	}

	friends := datastructures.FriendsFromDB{}
	err = json.Unmarshal(body, &friends)
	if err != nil {
		return datastructures.FriendsFromDB{}, err
	}

	return friends, nil
}
