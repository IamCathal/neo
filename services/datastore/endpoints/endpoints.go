package endpoints

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/IamCathal/neo/services/datastore/app"
	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/IamCathal/neo/services/datastore/dbmonitor"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/segmentio/ksuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	authRequiredEndpoints map[string]bool
)

type Endpoints struct {
	Cntr controller.CntrInterface
}

func init() {
	authRequiredEndpoints = make(map[string]bool)
	authRequiredEndpoints["saveuser"] = true
	authRequiredEndpoints["insertgame"] = true
	authRequiredEndpoints["getuser"] = true
	authRequiredEndpoints["getdetailsforgames"] = true
	authRequiredEndpoints["savecrawlingstats"] = true
	authRequiredEndpoints["getgraphabledata"] = true
	authRequiredEndpoints["getusernamesfromsteamids"] = true
	authRequiredEndpoints["saveprocessedgraphdata"] = true
}
func (endpoints *Endpoints) SetupRouter() *mux.Router {
	r := mux.NewRouter()

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/status", endpoints.Status).Methods("POST")
	apiRouter.HandleFunc("/saveuser", endpoints.SaveUser).Methods("POST")
	apiRouter.HandleFunc("/insertgame", endpoints.InsertGame).Methods("POST")
	apiRouter.HandleFunc("/getuser/{steamid}", endpoints.GetUser).Methods("GET")
	apiRouter.HandleFunc("/getdetailsforgames", endpoints.GetDetailsForGames).Methods("POST")
	apiRouter.HandleFunc("/savecrawlingstats", endpoints.SaveCrawlingStatsToDB).Methods("POST")
	apiRouter.HandleFunc("/getcrawlinguser/{crawlid}", endpoints.GetCrawlingUser).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/hasbeencrawledbefore", endpoints.HasBeenCrawledBefore).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/getcrawlingstatus/{crawlid}", endpoints.GetCrawlingStatus).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/getgraphabledata/{steamid}", endpoints.GetGraphableData).Methods("GET")
	apiRouter.HandleFunc("/getusernamesfromsteamids", endpoints.GetUsernamesFromSteamIDs).Methods("POST")
	apiRouter.HandleFunc("/saveprocessedgraphdata/{crawlid}", endpoints.SaveProcessedGraphData).Methods("POST")
	apiRouter.HandleFunc("/getprocessedgraphdata/{crawlid}", endpoints.GetProcessedGraphData).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/doesprocessedgraphdataexist/{crawlid}", endpoints.DoesProcessedGraphDataExist).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/calculateshortestdistanceinfo", endpoints.CalculateShortestDistanceInfo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/getshortestdistanceinfo", endpoints.GetShortestDistanceInfo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/getfinishedcrawlsaftertimestamp", endpoints.GetFinishedCrawlsAfterTimestamp).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/getfinishedshortestdistancecrawlsaftertimestamp", endpoints.GetFinishedShortestDistanceCrawlsAfterTimestamp).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/gettotalusersindb", endpoints.GetTotalUsersInDB).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/gettotalcrawlscompleted", endpoints.GetTotalCrawlsCompleted).Methods("GET", "OPTIONS")
	apiRouter.Use(endpoints.AuthMiddleware)
	apiRouter.Use(endpoints.LoggingMiddleware)

	wsRouter := r.PathPrefix("/ws").Subrouter()
	wsRouter.HandleFunc("/newuserstream", endpoints.NewUserStream).Methods("GET")
	wsRouter.HandleFunc("/crawlingstatstream/{crawlid}", endpoints.CrawlingStatsUpdateStream).Methods("GET")

	return r
}

func (endpoints *Endpoints) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseURLPath := util.GetBaseURLPath(r)
		if _, requiresAuth := authRequiredEndpoints[baseURLPath]; requiresAuth {
			if r.Header.Get("Authentication") != os.Getenv("AUTH_KEY") {

				configuration.Logger.Sugar().Infof("ip: %s with user-agent: %s wasn't authorized to access %s",
					r.RemoteAddr, r.Header.Get("User-Agent"), r.URL.Path)

				w.WriteHeader(http.StatusForbidden)
				response := struct {
					Error string `json:"error"`
				}{
					"You are not authorized to access this endpoint",
				}
				json.NewEncoder(w).Encode(response)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (endpoints *Endpoints) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		util.SetupCORS(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		defer func() {
			if err := recover(); err != nil {
				vars := mux.Vars(r)
				w.WriteHeader(http.StatusInternalServerError)
				response := struct {
					Error string `json:"error"`
				}{
					fmt.Sprintf("Give the code monkeys this ID: '%s'", vars["requestID"]),
				}
				json.NewEncoder(w).Encode(response)

				requestStartTime, timeParseErr := strconv.ParseInt(vars["requestStartTime"], 10, 64)
				if timeParseErr != nil {
					configuration.Logger.Fatal(fmt.Sprintf("%v", err),
						zap.String("requestID", vars["requestID"]),
						zap.Int("status", http.StatusInternalServerError),
						zap.Int64("duration", util.GetCurrentTimeInMs()-requestStartTime),
						zap.String("path", r.URL.EscapedPath()),
						zap.String("shortPath", util.GetBaseURLPath(r)),
					)
					panic(timeParseErr)
				}

				configuration.Logger.Error(fmt.Sprintf("%v", err),
					zap.String("requestID", vars["requestID"]),
					zap.Int("status", http.StatusInternalServerError),
					zap.Int64("duration", util.GetCurrentTimeInMs()-requestStartTime),
					zap.String("path", r.URL.EscapedPath()),
					zap.String("shortPath", util.GetBaseURLPath(r)),
				)
			}
		}()

		vars := mux.Vars(r)

		identifier := ksuid.New()
		vars["requestID"] = identifier.String()
		requestStartTime := time.Now().UnixNano() / int64(time.Millisecond)
		vars["requestStartTime"] = strconv.Itoa(int(requestStartTime))

		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)

		configuration.Logger.Info("served content",
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", wrapped.status),
			zap.Int64("duration", util.GetCurrentTimeInMs()-requestStartTime),
			zap.String("path", r.URL.EscapedPath()),
			zap.String("shortPath", util.GetBaseURLPath(r)),
		)

		writeAPI := configuration.InfluxDBClient.WriteAPI(os.Getenv("ORG"), os.Getenv("ENDPOINT_LATENCIES_BUCKET"))
		point := influxdb2.NewPointWithMeasurement("endpointLatencies").
			AddTag("path", util.GetBaseURLPath(r)).
			AddTag("service", "datastore").
			AddField("latency", util.GetCurrentTimeInMs()-requestStartTime).
			SetTime(time.Now())
		writeAPI.WritePoint(point)
	})
}

func (endpoints *Endpoints) SaveUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	saveUserDTO := dtos.SaveUserDTO{}

	err := json.NewDecoder(r.Body).Decode(&saveUserDTO)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	crawlingStats := common.CrawlingStatus{
		CrawlID:             saveUserDTO.CrawlID,
		OriginalCrawlTarget: saveUserDTO.OriginalCrawlTarget,
		MaxLevel:            saveUserDTO.MaxLevel,
		TotalUsersToCrawl:   len(saveUserDTO.User.FriendIDs),
	}

	err = app.SaveCrawlingStatsToDB(endpoints.Cntr, saveUserDTO.CurrentLevel, crawlingStats)
	if err != nil {
		logMsg := fmt.Sprintf("failed to save crawling stats to DB: %+v", err)
		configuration.Logger.Sugar().Error(logMsg)
		panic(logMsg)
	}

	err = app.SaveUserToDB(endpoints.Cntr, saveUserDTO.User)
	if err != nil {
		logMsg := fmt.Sprintf("failed to save user to DB: %+v", err)
		configuration.Logger.Error(logMsg)
		panic(logMsg)
	}

	configuration.Logger.Sugar().Infof("successfully saved user %s to db", saveUserDTO.User.AccDetails.SteamID)

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		"success",
		"very good",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) InsertGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bareGameInfo := common.BareGameInfo{}

	err := json.NewDecoder(r.Body).Decode(&bareGameInfo)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	success, err := endpoints.Cntr.InsertGame(context.TODO(), bareGameInfo)
	if err != nil || !success {
		logMsg := fmt.Sprintf("failed to insert game: %+v", err)
		configuration.Logger.Sugar().Error(logMsg)
		panic(logMsg)
	}

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		"success",
		"very good",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) SaveCrawlingStatsToDB(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	crawlingStatusInput := dtos.SaveCrawlingStatsDTO{}

	err := json.NewDecoder(r.Body).Decode(&crawlingStatusInput)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	err = app.SaveCrawlingStatsToDB(endpoints.Cntr, crawlingStatusInput.CurrentLevel, crawlingStatusInput.CrawlingStatus)
	if err != nil {
		logMsg := fmt.Sprintf("failed to save crawling stats: %+v", err)
		configuration.Logger.Sugar().Error(logMsg)
		panic(logMsg)
	}

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		"success",
		"very good",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) GetCrawlingUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Validate crawlid
	_, err := ksuid.Parse(vars["crawlid"])
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "invalid crawlid", vars, http.StatusNotFound)
		return
	}
	// If the user is a crawlTarget and crawling is in progress
	isCurrentlyBeingCrawled, steamID, err := app.IsCurrentlyBeingCrawled(endpoints.Cntr, vars["crawlid"])
	if err != nil {
		logMsg := fmt.Sprintf("could not check crawling progress: %+v", err)
		configuration.Logger.Sugar().Error(logMsg)
		panic(logMsg)
	}
	if !isCurrentlyBeingCrawled {
		util.SendBasicInvalidResponse(w, r, "User is not currently being crawled", vars, http.StatusNotFound)
		return
	}

	user, err := app.GetUserFromDB(endpoints.Cntr, steamID)
	if err == mongo.ErrNoDocuments {
		util.SendBasicInvalidResponse(w, r, "user does not exist", vars, http.StatusNotFound)
		return
	}
	if err != nil {
		logMsg := fmt.Sprintf("couldn't get user: %+v", err)
		configuration.Logger.Sugar().Error(logMsg)
		panic(logMsg)
	}

	response := struct {
		Status string              `json:"status"`
		User   common.UserDocument `json:"user"`
	}{
		"success",
		user,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) HasBeenCrawledBefore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	crawlDetails := dtos.HasBeenCrawledBeforeInputDTO{}

	err := json.NewDecoder(r.Body).Decode(&crawlDetails)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	if isValidFormat := util.IsValidFormatSteamID(crawlDetails.SteamID); !isValidFormat {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}
	if crawlDetails.Level < 1 || crawlDetails.Level > 3 {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	crawlID, err := endpoints.Cntr.HasUserBeenCrawledBeforeAtLevel(context.TODO(), crawlDetails.Level, crawlDetails.SteamID)
	if err != nil {
		logMsg := fmt.Sprintf("couldn't lookup has user been crawled before: %+v", err)
		configuration.Logger.Error(logMsg)
		panic(err)
	}

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		"success",
		crawlID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Validate steamid
	if isValid := util.IsValidFormatSteamID(vars["steamid"]); !isValid {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	user, err := app.GetUserFromDB(endpoints.Cntr, vars["steamid"])
	if err == mongo.ErrNoDocuments {
		util.SendBasicInvalidResponse(w, r, "user does not exist", vars, http.StatusNotFound)
		return
	}
	if err != nil {
		logMsg := fmt.Sprintf("couldn't get user: %+v", err)
		configuration.Logger.Error(logMsg)
		panic(logMsg)
	}

	response := struct {
		Status string              `json:"status"`
		User   common.UserDocument `json:"user"`
	}{
		"success",
		user,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func (endpoints *Endpoints) GetDetailsForGames(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	gamesInput := dtos.GetDetailsForGamesInputDTO{}
	err := json.NewDecoder(r.Body).Decode(&gamesInput)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	if len(gamesInput.GameIDs) == 0 || len(gamesInput.GameIDs) > 20 {
		util.SendBasicInvalidResponse(w, r, "Can only request 1-20 games in a request", vars, http.StatusBadRequest)
		return
	}

	gameDetails, err := endpoints.Cntr.GetDetailsForGames(context.TODO(), gamesInput.GameIDs)
	if err != nil {
		logMsg := fmt.Sprintf("error retrieving games: %+v", err)
		configuration.Logger.Error(logMsg)
		panic(logMsg)
	}

	if len(gameDetails) == 0 {
		gameDetails = []common.BareGameInfo{}
	}

	response := dtos.GetDetailsForGamesDTO{
		Status: "success",
		Games:  gameDetails,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) GetCrawlingStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, err := ksuid.Parse(vars["crawlid"])
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "invalid crawlid", vars, http.StatusNotFound)
		return
	}
	crawlingStatus, err := app.GetCrawlingStatsFromDBFromCrawlID(endpoints.Cntr, vars["crawlid"])
	if err != nil {
		logMsg := fmt.Sprintf("couldn't get crawling status: %+v", err)
		configuration.Logger.Error(logMsg)
		panic(logMsg)
	}
	response := dtos.GetCrawlingStatusDTO{
		Status:         "success",
		CrawlingStatus: crawlingStatus,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) GetGraphableData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// Validate steamid
	if isValid := util.IsValidFormatSteamID(vars["steamid"]); !isValid {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	user, err := endpoints.Cntr.GetUser(context.TODO(), vars["steamid"])
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "couldn't get user", vars, http.StatusNotFound)
		return
	}

	graphableDataForUser := dtos.GetGraphableDataForUserDTO{
		Username:  user.AccDetails.Personaname,
		SteamID:   user.AccDetails.SteamID,
		FriendIDs: user.FriendIDs,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(graphableDataForUser)
}

func (endpoints *Endpoints) GetUsernamesFromSteamIDs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	steamIDsInput := dtos.GetUsernamesFromSteamIDsInputDTO{}

	err := json.NewDecoder(r.Body).Decode(&steamIDsInput)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	// Validate all given steamids
	for _, steamID := range steamIDsInput.SteamIDs {
		if isValid := util.IsValidFormatSteamID(steamID); !isValid {
			util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
			return
		}
	}

	steamIDsToUsernames, err := endpoints.Cntr.GetUsernames(context.TODO(), steamIDsInput.SteamIDs)
	if err != nil {
		configuration.Logger.Sugar().Errorf("couldn't get usernames: %+v", err)
		util.SendBasicInvalidResponse(w, r, "couldn't get usernames", vars, http.StatusBadRequest)
		return

	}

	returnJSON := dtos.GetUsernamesFromSteamIDsDTO{}
	for key, val := range steamIDsToUsernames {
		currentMapping := dtos.SteamIDAndUsername{
			SteamID:  key,
			Username: val,
		}
		returnJSON.SteamIDAndUsername = append(returnJSON.SteamIDAndUsername, currentMapping)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(returnJSON)
}

func (endpoints *Endpoints) SaveProcessedGraphData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, err := ksuid.Parse(vars["crawlid"])
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "invalid input", vars, http.StatusBadRequest)
		return
	}

	graphData := common.UsersGraphData{}
	reqBodyBytes, err := gunzip(r.Body)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "invalid input", vars, http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(reqBodyBytes, &graphData)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	success, err := endpoints.Cntr.SaveProcessedGraphData(vars["crawlid"], graphData)
	if err != nil || !success {
		configuration.Logger.Sugar().Errorf("could not save graph data: %+v", err)
		util.SendBasicInvalidResponse(w, r, "could not save graph data", vars, http.StatusBadRequest)
		return
	}

	response := struct {
		Status string `json:"status"`
	}{
		"success",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) GetProcessedGraphData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, err := ksuid.Parse(vars["crawlid"])
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "invalid input", vars, http.StatusBadRequest)
		return
	}
	responseType := r.URL.Query().Get("responsetype")
	if responseType != "json" {
		if responseType != "" {
			util.SendBasicInvalidResponse(w, r, "invalid input", vars, http.StatusBadRequest)
			return
		}
	}

	usersProcessedGraphData, err := endpoints.Cntr.GetProcessedGraphData(vars["crawlid"])
	if err != nil {
		configuration.Logger.Sugar().Errorf("failed to get processed graph data: %+v", err)
		util.SendBasicInvalidResponse(w, r, "failed to get processed graph data", vars, http.StatusBadRequest)
		return
	}
	response := datastructures.GetProcessedGraphDataDTO{
		Status:        "success",
		UserGraphData: usersProcessedGraphData,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		configuration.Logger.Sugar().Errorf("failed to marshal processedgraphdata: %+v", err)
	}

	switch responseType {
	case "json":
		w.Header().Set("Content-Length", strconv.Itoa(len(string(jsonResponse))+1))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	default:
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "application/javascript")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gz.Write(jsonResponse)
	}
}

func (endpoints *Endpoints) DoesProcessedGraphDataExist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, err := ksuid.Parse(vars["crawlid"])
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "invalid input", vars, http.StatusBadRequest)
		return
	}

	exists, err := endpoints.Cntr.DoesProcessedGraphDataExist(vars["crawlid"])
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "failed to get processed graph data", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Errorf("failed to get processed graph data: %+v", err)
		return
	}

	response := dtos.DoesProcessedGraphDataExistDTO{
		Status: "success",
	}
	if exists {
		response.Exists = "yes"
	} else {
		response.Exists = "no"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) CalculateShortestDistanceInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	crawlIDsInput := datastructures.GetShortestDistanceInfoDataInputDTO{}
	err := json.NewDecoder(r.Body).Decode(&crawlIDsInput)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}

	for _, crawlID := range crawlIDsInput.CrawlIDs {
		_, err := ksuid.Parse(crawlID)
		if err != nil {
			util.SendBasicInvalidResponse(w, r, "invalid crawlid", vars, http.StatusBadRequest)
			return
		}
	}
	if len(crawlIDsInput.CrawlIDs) != 2 {
		util.SendBasicInvalidResponse(w, r, "two crawl IDS must be given", vars, http.StatusBadRequest)
		return
	}

	exists, existingShortestDistanceInfo, err := endpoints.Cntr.GetShortestDistanceInfo(context.TODO(), crawlIDsInput.CrawlIDs)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "could not find shortest distance", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Errorf("failed to retrieve existing shortestDistanceInfo: %+v", err)
	}
	if exists {
		response := struct {
			Status string                              `json:"status"`
			Data   datastructures.ShortestDistanceInfo `json:"shortestdistanceinfo"`
		}{
			"success",
			existingShortestDistanceInfo,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	_, shortestDistanceInfo, err := app.CalulateShortestDistanceInfo(endpoints.Cntr, crawlIDsInput.CrawlIDs[0], crawlIDsInput.CrawlIDs[1])
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "could not find shortest distance", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Errorf("failed to get shortest distance: %s", err.Error())
		return
	}

	success, err := endpoints.Cntr.SaveShortestDistance(context.TODO(), shortestDistanceInfo)
	if err != nil || !success {
		util.SendBasicInvalidResponse(w, r, "could not save shortest distance", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Errorf("could not save shortest distance: %s", err.Error())
		return
	}

	response := struct {
		Status string                              `json:"status"`
		Data   datastructures.ShortestDistanceInfo `json:"shortestdistanceinfo"`
	}{
		"success",
		shortestDistanceInfo,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func (endpoints *Endpoints) GetShortestDistanceInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	crawlIDsInput := datastructures.GetShortestDistanceInfoDataInputDTO{}
	err := json.NewDecoder(r.Body).Decode(&crawlIDsInput)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}
	for _, crawlID := range crawlIDsInput.CrawlIDs {
		_, err := ksuid.Parse(crawlID)
		if err != nil {
			util.SendBasicInvalidResponse(w, r, "invalid crawlid", vars, http.StatusBadRequest)
			return
		}
	}
	if len(crawlIDsInput.CrawlIDs) != 2 {
		util.SendBasicInvalidResponse(w, r, "two crawl IDS must be given", vars, http.StatusBadRequest)
		return
	}

	_, shortestDistanceInfo, err := endpoints.Cntr.GetShortestDistanceInfo(context.TODO(), crawlIDsInput.CrawlIDs)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "could not get shortest distance", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Errorf("could not get shortest distance: %s", err.Error())
		return
	}

	response := struct {
		Status string                              `json:"status"`
		Data   datastructures.ShortestDistanceInfo `json:"shortestdistanceinfo"`
	}{
		"success",
		shortestDistanceInfo,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) GetFinishedCrawlsAfterTimestamp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	timestampString := r.URL.Query().Get("timestamp")
	timeStampint64, err := strconv.ParseInt(timestampString, 10, 64)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "invalid timestamp", vars, http.StatusBadRequest)
		return
	}

	crawlsFinishedAfterTimeStamp := dbmonitor.GetRecentFinishedCrawlsAfterTimestamp(timeStampint64)

	response := datastructures.GetFinishedCrawlsDTO{
		Status:                     "success",
		AllFinishedCrawlsWithUsers: crawlsFinishedAfterTimeStamp,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) GetFinishedShortestDistanceCrawlsAfterTimestamp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	timestampString := r.URL.Query().Get("timestamp")
	timeStampint64, err := strconv.ParseInt(timestampString, 10, 64)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "invalid timestamp", vars, http.StatusBadRequest)
		return
	}

	crawlsFinishedAfterTimeStamp := dbmonitor.GetRecentFinishedShortestDistanceCrawlsAfterTimestamp(timeStampint64)

	response := datastructures.GetFinishedShortestDistanceCrawlsDTO{
		Status:         "success",
		CrawlingStatus: crawlsFinishedAfterTimeStamp,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) GetTotalUsersInDB(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Status    string `json:"status"`
		Usersindb int64  `json:"usersindb"`
	}{
		"success",
		dbmonitor.GetTotalUsersInDB(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) GetTotalCrawlsCompleted(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Status      string `json:"status"`
		Totalcrawls int64  `json:"totalcrawls"`
	}{
		"success",
		dbmonitor.GetTotalCrawlsCompleted(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (endpoints *Endpoints) Status(w http.ResponseWriter, r *http.Request) {
	req := common.UptimeResponse{
		Uptime: time.Since(configuration.ApplicationStartUpTime),
		Status: "operational",
	}
	jsonObj, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}
