package endpoints

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/IamCathal/neo/services/datastore/app"
	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/gorilla/mux"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/segmentio/ksuid"
	"go.mongodb.org/mongo-driver/mongo"
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
	r.HandleFunc("/saveuser", endpoints.SaveUser).Methods("POST")
	r.HandleFunc("/getuser/{steamid}", endpoints.GetUser).Methods("GET")

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
				fmt.Println(err)
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
		// Authenticate JWT
		next.ServeHTTP(w, r)
	})
}

func (endpoints *Endpoints) SaveUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	saveUserDTO := dtos.SaveUserDTO{}

	err := json.NewDecoder(r.Body).Decode(&saveUserDTO)
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		LogBasicErr(err, r, http.StatusBadRequest)
		return
	}
	err = app.SaveCrawlingStatsToDB(endpoints.Cntr, saveUserDTO)
	if err != nil {
		LogBasicErr(err, r, http.StatusBadRequest)
		util.SendBasicInvalidResponse(w, r, "cannot save crawling stats", vars, http.StatusBadRequest)
		return
	}
	err = app.SaveUserToDB(endpoints.Cntr, saveUserDTO.User)
	if err != nil {
		LogBasicErr(err, r, http.StatusBadRequest)
		util.SendBasicInvalidResponse(w, r, "cannot save user", vars, http.StatusBadRequest)
		return
	}
	configuration.Logger.Sugar().Infof("successfully saved user %s to db", saveUserDTO.User.SteamID)

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

func (endpoints *Endpoints) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Validate steamid
	if isValid := util.IsValidFormatSteamID(vars["steamid"]); !isValid {
		util.SendBasicInvalidResponse(w, r, "Invalid input", vars, http.StatusBadRequest)
		LogBasicInfo("invalid steamID given", r, http.StatusBadRequest)
		return
	}

	user, err := app.GetUserFromDB(endpoints.Cntr, vars["steamid"])
	if err == mongo.ErrNoDocuments {
		util.SendBasicInvalidResponse(w, r, "user does not exist", vars, http.StatusNotFound)
		LogBasicInfo("user was not found in DB", r, http.StatusBadRequest)
		return
	}

	if err != nil {
		util.SendBasicInvalidResponse(w, r, "couldn't get user", vars, http.StatusBadRequest)
		LogBasicInfo("couldn't get user", r, http.StatusBadRequest)
		return
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
