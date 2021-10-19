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
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/iamcathal/neo/services/crawler/util"
	"github.com/iamcathal/neo/services/crawler/worker"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type Endpoints struct {
	ApplicationStartUpTime time.Time
	Logger                 *zap.Logger
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
// Taken from https://blog.questionable.services/article/guide-logging-middleware-go/
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/status", Status).Methods("POST")
	r.HandleFunc("/crawl", CrawlUsers).Methods("POST")

	r.Use(LoggingMiddleware)
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

func LoggingMiddleware(next http.Handler) http.Handler {
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
					util.LogBasicFatal(timeParseErr, vars, r, http.StatusInternalServerError)
					panic(timeParseErr)
				}

				util.LogBasicErr(errors.New(fmt.Sprintf("%v", err)), vars, r, http.StatusInternalServerError)
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
			zap.Int64("duration", configuration.GetCurrentTimeInMs()-requestStartTime),
			zap.String("path", r.URL.EscapedPath()),
		)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func Status(w http.ResponseWriter, r *http.Request) {
	req := datastructures.UptimeResponse{
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

func CrawlUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userInput := datastructures.CrawlUsersInput{}

	err := json.NewDecoder(r.Body).Decode(&userInput)
	if err != nil {
		util.SendBasicErrorResponse(w, r, err, vars, http.StatusBadRequest)
		util.LogBasicErr(err, vars, r, http.StatusBadRequest)
		return
	}

	validSteamIDs, err := worker.VerifyFormatOfSteamIDs(userInput)
	if err != nil {
		util.SendBasicErrorResponse(w, r, err, vars, http.StatusBadRequest)
		util.LogBasicErr(err, vars, r, http.StatusBadRequest)
		return
	}
	if len(validSteamIDs) == 0 {
		util.SendBasicInvalidResponse(w, r, "No valid format steamIDs sent", vars, http.StatusBadRequest)
	}

	util.LogBasicInfo(fmt.Sprintf("received valid format steamIDs: %+v", validSteamIDs), vars, r, http.StatusOK)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}
