package endpoints

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/IamCathal/neo/services/frontend/app"
	"github.com/IamCathal/neo/services/frontend/configuration"
	"github.com/gorilla/mux"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type Endpoints struct {
	ApplicationStartUpTime time.Time
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
	r.HandleFunc("/", endpoints.HomeHandler).Methods("GET")
	r.HandleFunc("/graph/{graphID}", endpoints.ServeGraph).Methods("GET")
	r.HandleFunc("/status", endpoints.Status).Methods("POST")
	r.HandleFunc("/isprivateprofile/{steamid}", endpoints.IsPrivateProfile).Methods("GET")
	r.HandleFunc("/createcrawlingstatus", endpoints.CreateCrawlingStatus).Methods("POST")

	r.Use(endpoints.LoggingMiddleware)

	r.Handle("/static", http.NotFoundHandler())
	fs := http.FileServer(http.Dir(os.Getenv("STATIC_CONTENT_DIR")))
	r.PathPrefix("/").Handler(endpoints.DisallowFileBrowsing(fs))

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

func (endpoints *Endpoints) DisallowFileBrowsing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
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

				requestStartTime, timeParseErr := strconv.ParseInt(vars["requestStartTime"], 10, 64)
				if timeParseErr != nil {
					configuration.Logger.Fatal(fmt.Sprintf("%v", err),
						zap.String("requestID", vars["requestID"]),
						zap.Int("status", http.StatusInternalServerError),
						zap.Int64("duration", util.GetCurrentTimeInMs()-requestStartTime),
						zap.String("path", r.URL.EscapedPath()),
					)
					panic(timeParseErr)
				}

				configuration.Logger.Error(fmt.Sprintf("%v", err),
					zap.String("requestID", vars["requestID"]),
					zap.Int("status", http.StatusInternalServerError),
					zap.Int64("duration", util.GetCurrentTimeInMs()-requestStartTime),
					zap.String("path", r.URL.EscapedPath()),
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
		)
	})
}

func (endpoints *Endpoints) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (endpoints *Endpoints) HomeHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, fmt.Sprintf("%s/pages/index.html", os.Getenv("STATIC_CONTENT_DIR_NAME")))
}

func (endpoints *Endpoints) ServeGraph(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	isValidFormatGraphID, err := util.IsValidFormatGraphID(path.Clean(vars["graphID"]))
	if err != nil || !isValidFormatGraphID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := struct {
			Error string `json:"error"`
		}{
			"invalid graphID given",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	http.ServeFile(w, req, fmt.Sprintf("%s/pages/%s.html", os.Getenv("STATIC_CONTENT_DIR_NAME"), vars["graphID"]))
}

func (endpoints *Endpoints) IsPrivateProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if isValid := util.IsValidFormatSteamID(vars["steamid"]); isValid == false {
		util.SendBasicInvalidResponse(w, r, "invalid steamid given", vars, http.StatusBadRequest)
		return
	}

	res, err := util.GetAndRead(
		fmt.Sprintf("%s/isprivateprofile/%s", os.Getenv("CRAWLER_INSTANCE"), vars["steamid"]))
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "could not check status of steam profile", vars, http.StatusBadRequest)
		configuration.Logger.Sugar().Warnf("failed to call isprivateprofile for %s: %v", vars["steamid"], err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(res))
}

func (endpoints *Endpoints) CreateCrawlingStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	crawlingStatus := dtos.SaveCrawlingStatsDTO{}

	err := json.NewDecoder(r.Body).Decode(&crawlingStatus)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		LogBasicErr(err, r, http.StatusBadRequest)
		return
	}

	success, err := app.CreateCrawlingStatus(crawlingStatus)
	if err != nil || success == false {
		util.SendBasicInvalidResponse(w, r, "Error saving crawling status", vars, http.StatusBadRequest)
		LogBasicErr(err, r, http.StatusBadRequest)
		return
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

func (endpoints *Endpoints) Status(w http.ResponseWriter, r *http.Request) {
	req := common.UptimeResponse{
		Uptime: time.Since(endpoints.ApplicationStartUpTime),
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
