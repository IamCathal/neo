package endpoints

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/iamcathal/neo/services/crawler/graphing"
	"github.com/iamcathal/neo/services/crawler/worker"
	"github.com/neosteamfriendgraphing/common"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type Endpoints struct {
	Cntr controller.CntrInterface
}

func (endpoints *Endpoints) SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/status", endpoints.Status).Methods("POST")
	r.HandleFunc("/crawl", endpoints.CrawlUsers).Methods("POST", "OPTIONS")
	r.HandleFunc("/isprivateprofile/{steamid}", endpoints.IsPrivateProfile).Methods("GET", "OPTIONS")
	r.HandleFunc("/creategraph/{crawlid}", endpoints.CreateGraph).Methods("POST", "OPTIONS")

	r.Use(endpoints.LoggingMiddleware)
	return r
}

func (endpoints *Endpoints) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commonUtil.SetupCORS(&w, r)
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

				_, timeParseErr := strconv.ParseInt(vars["requestStartTime"], 10, 64)
				if timeParseErr != nil {
					configuration.Logger.Sugar().Errorf("failed to parse time in middleware: %+v", commonUtil.MakeErr(timeParseErr))
					panic(timeParseErr)
				}
				configuration.Logger.Sugar().Errorf("panic caught in middleware: %+v", err)
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
			zap.Int64("duration", commonUtil.GetCurrentTimeInMs()-requestStartTime),
			zap.String("path", r.URL.EscapedPath()),
		)
	})
}

func (endpoints *Endpoints) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (endpoints *Endpoints) Status(w http.ResponseWriter, r *http.Request) {
	req := common.UptimeResponse{
		Uptime: time.Since(configuration.ApplicationStartUpTime),
		Status: "operational",
	}
	jsonObj, err := json.Marshal(req)
	if err != nil {
		log.Fatal(commonUtil.MakeErr(err))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}

func (endpoints *Endpoints) CrawlUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	userInput := datastructures.CrawlUserTempDTO{}
	err := json.NewDecoder(r.Body).Decode(&userInput)
	if err != nil {
		commonUtil.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		return
	}
	if userInput.Level < 2 || userInput.Level > 3 {
		commonUtil.SendBasicInvalidResponse(w, r, "Invalid level given", vars, http.StatusBadRequest)
		return
	}
	if len(userInput.SteamIDs) == 0 {
		commonUtil.SendBasicInvalidResponse(w, r, "No steamIDs given", vars, http.StatusBadRequest)
		return
	}
	for _, steamID := range userInput.SteamIDs {
		if isValid := commonUtil.IsValidFormatSteamID(steamID); !isValid {
			commonUtil.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
			return
		}
	}

	firstCrawlID := vars["requestID"]
	secondCrawlID := ""
	crawlIDsGenerated := []string{firstCrawlID}

	err = worker.CrawlUser(endpoints.Cntr, userInput.SteamIDs[0], firstCrawlID, userInput.Level)
	if err != nil {
		commonUtil.SendBasicInvalidResponse(w, r, "couldn't start crawl", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Errorf("failed to start crawl for first user with crawlID: %s steamID: %s level: %d", firstCrawlID, userInput.SteamIDs[0], userInput.Level)
		return
	}

	if len(userInput.SteamIDs) == 2 {
		secondCrawlID = ksuid.New().String()
		configuration.Logger.Sugar().Infof("creating new crawlID %s from request %s for user %s", secondCrawlID, firstCrawlID, userInput.SteamIDs[1])
		crawlIDsGenerated = append(crawlIDsGenerated, secondCrawlID)

		err = worker.CrawlUser(endpoints.Cntr, userInput.SteamIDs[1], secondCrawlID, userInput.Level)
		if err != nil {
			commonUtil.SendBasicInvalidResponse(w, r, "couldn't start crawl", vars, http.StatusBadRequest)
			configuration.Logger.Sugar().Errorf("failed to start crawl for second user with crawlID: %s steamID: %s level: %d", secondCrawlID, userInput.SteamIDs[1], userInput.Level)
			return
		}
	}

	response := datastructures.CrawlResponseDTO{
		Status:   "success",
		CrawlIDs: crawlIDsGenerated,
	}
	jsonObj, err := json.Marshal(response)
	if err != nil {
		commonUtil.SendBasicInvalidResponse(w, r, "couldn't return response", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Errorf("failed to marshal crawlResponse: %+v", commonUtil.MakeErr(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}

func (endpoints *Endpoints) IsPrivateProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if isValid := commonUtil.IsValidFormatSteamID(vars["steamid"]); !isValid {
		commonUtil.SendBasicInvalidResponse(w, r, "invalid steamid given", vars, http.StatusBadRequest)
		return
	}

	friends, err := endpoints.Cntr.CallGetFriends(vars["steamid"])
	if err != nil {
		commonUtil.SendBasicInvalidResponse(w, r, "invalid steamid given", vars, http.StatusBadRequest)
		return
	}

	response := common.BasicAPIResponse{
		Status: "success",
	}
	if len(friends) == 0 {
		// If the user has no friends they might have a public
		// account but we might as well consider them private
		// as we cannot crawl them
		response.Message = "private"
	} else {
		response.Message = "public"
	}

	jsonObj, err := json.Marshal(response)
	if err != nil {
		commonUtil.SendBasicInvalidResponse(w, r, "couldn't return response", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Errorf("failed to marshal BasicAPIResponse: %+v", commonUtil.MakeErr(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}

func (endpoints *Endpoints) CreateGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, err := ksuid.Parse(vars["crawlid"])
	if err != nil {
		commonUtil.SendBasicInvalidResponse(w, r, "invalid crawlid", vars, http.StatusBadRequest)
		return
	}

	// Check if this crawl session is actually finished
	crawlingStats, err := endpoints.Cntr.GetCrawlingStatsFromDataStore(vars["crawlid"])
	if err != nil {
		commonUtil.SendBasicInvalidResponse(w, r, "could not check if crawling has finished", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Errorf("failed to retrieve crawling status: %+v", err)
		return
	}
	graphWorkerConfig := graphing.GraphWorkerConfig{
		TotalUsersToCrawl: crawlingStats.TotalUsersToCrawl,
		UsersCrawled:      0,
		MaxLevel:          crawlingStats.MaxLevel,
	}

	go graphing.CollectGraphData(endpoints.Cntr, crawlingStats.OriginalCrawlTarget, vars["crawlid"], graphWorkerConfig)

	response := common.BasicAPIResponse{
		Status:  "success",
		Message: "graph creation has been initiated",
	}
	jsonObj, err := json.Marshal(response)
	if err != nil {
		commonUtil.SendBasicInvalidResponse(w, r, "couldn't return response", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Errorf("failed to marshal BasicAPIResponse: %+v", commonUtil.MakeErr(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}
