package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/iamcathal/neo/services/crawler/util"
	"github.com/iamcathal/neo/services/crawler/worker"
	"github.com/neosteamfriendgraphing/common"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type Endpoints struct {
	Cntr controller.CntrInterface
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
// Taken from https://blog.questionable.services/article/guide-logging-middleware-go/
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (endpoints *Endpoints) SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/status", endpoints.Status).Methods("POST")
	r.HandleFunc("/crawl", endpoints.CrawlUsers).Methods("POST")
	r.HandleFunc("/isprivateprofile/{steamid}", endpoints.IsPrivateProfile).Methods("GET")

	r.Use(endpoints.LoggingMiddleware)
	return r
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (endpoints *Endpoints) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
					util.LogBasicFatal(timeParseErr, r, http.StatusInternalServerError)
					panic(timeParseErr)
				}

				util.LogBasicErr(errors.New(fmt.Sprintf("%v", err)), r, http.StatusInternalServerError)
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
		log.Fatal(util.MakeErr(err))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}

func (endpoints *Endpoints) CrawlUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	userInput := datastructures.CrawlUsersInput{}
	err := json.NewDecoder(r.Body).Decode(&userInput)
	if err != nil {
		commonUtil.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		util.LogBasicErr(err, r, http.StatusBadRequest)
		return
	}
	if userInput.Level < 1 || userInput.Level > 3 {
		commonUtil.SendBasicInvalidResponse(w, r, "Invalid level given", vars, http.StatusBadRequest)
		return
	}

	validSteamIDs, err := worker.VerifyFormatOfSteamIDs(userInput)
	if err != nil {
		commonUtil.SendBasicErrorResponse(w, r, err, vars, http.StatusBadRequest)
		util.LogBasicErr(err, r, http.StatusBadRequest)
		return
	}
	if len(validSteamIDs) == 0 {
		commonUtil.SendBasicInvalidResponse(w, r, "No valid format steamIDs sent", vars, http.StatusBadRequest)
		return
	}
	util.LogBasicInfo(fmt.Sprintf("received valid format steamIDs: %+v with level: %d", validSteamIDs, userInput.Level), r, http.StatusOK)

	// TODO make a new crawlID when crawling a second user and log it to see the connection
	// between requestID and this new crawlID
	worker.CrawlUser(endpoints.Cntr, validSteamIDs[0], vars["requestID"], userInput.Level)
	// if err != nil {
	// 	util.SendBasicErrorResponse(w, r, err, vars, http.StatusBadRequest)
	// 	util.LogBasicErr(err, vars, r, http.StatusBadRequest)
	// 	return
	// }
	response := common.BasicAPIResponse{
		Status:  "success",
		Message: vars["requestID"],
	}
	jsonObj, err := json.Marshal(response)
	if err != nil {
		log.Fatal(util.MakeErr(err))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}

func (endpoints *Endpoints) IsPrivateProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if isValid := commonUtil.IsValidFormatSteamID(vars["steamid"]); isValid == false {
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
		response.Message = "private"
	} else {
		response.Message = "public"
	}

	jsonObj, err := json.Marshal(response)
	if err != nil {
		log.Fatal(util.MakeErr(err))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))

}
