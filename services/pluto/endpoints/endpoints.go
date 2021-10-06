package endpoints

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IamCathal/neo/services/pluto/datastructures"
	"github.com/IamCathal/neo/services/pluto/util"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type Endpoints struct {
	ApplicationStartUpTime time.Time
	ChannelID              string
	Logger                 *zap.Logger
	DiscordSession         *discordgo.Session
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
// Taken from https://blog.questionable.services/article/guide-logging-middleware-go/
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
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

				requestStartTime, timeParseErr := strconv.ParseInt(vars["requestStartTime"], 10, 64)
				if timeParseErr != nil {
					endpoints.Logger.Fatal(fmt.Sprintf("%v", err),
						zap.String("requestID", vars["requestID"]),
						zap.Int("status", http.StatusInternalServerError),
						zap.Int64("duration", util.GetCurrentTimeInMs()-requestStartTime),
						zap.String("path", r.URL.EscapedPath()),
					)
					panic(timeParseErr)
				}

				endpoints.Logger.Error(fmt.Sprintf("%v", err),
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

		endpoints.Logger.Info("served content",
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", wrapped.status),
			zap.Int64("duration", util.GetCurrentTimeInMs()-requestStartTime),
			zap.String("path", r.URL.EscapedPath()),
		)
	})
}

func (endpoints *Endpoints) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	gotHash := strings.SplitN(r.Header.Get("X-Hub-Signature-256"), "=", 2)
	if gotHash[0] != "sha256" {
		endpoints.Logger.Error(fmt.Sprintf("invalid format signature '%s' received", gotHash[0]),
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", http.StatusInternalServerError),
			zap.String("path", r.URL.EscapedPath()),
		)
		panic("invalid signature format received")
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		endpoints.Logger.Error(fmt.Sprintf("Cannot read the request body: %s\n", err),
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", http.StatusInternalServerError),
			zap.String("path", r.URL.EscapedPath()),
		)
		panic("cannot read request body")
	}
	hash := hmac.New(sha256.New, []byte(os.Getenv("WEBHOOK_SECRET")))
	if _, err := hash.Write(b); err != nil {
		endpoints.Logger.Error(fmt.Sprintf("Cannot compute the HMAC for request: %s\n", err),
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", http.StatusInternalServerError),
			zap.String("path", r.URL.EscapedPath()),
		)
		panic("cannot compute HMAC")
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))
	if gotHash[1] != expectedHash {
		endpoints.Logger.Error(fmt.Sprintf("invalid signature '%s' does not match expected '%s'", gotHash[1], expectedHash),
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", http.StatusInternalServerError),
			zap.String("path", r.URL.EscapedPath()),
		)
		panic("invalid secret given")
	}

	var reqData datastructures.WebhookPayload
	err = json.Unmarshal(b, &reqData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := struct {
			Error string `json:"error"`
		}{
			"invalid input",
		}
		json.NewEncoder(w).Encode(response)
		endpoints.Logger.Error(fmt.Sprintf("could not unmarshal request body: %s", err),
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", http.StatusInternalServerError),
			zap.String("path", r.URL.EscapedPath()),
		)
		return
	}

	endpoints.Logger.Info(fmt.Sprintf("%s was run with state: %s", reqData.WorkflowJob.Name, reqData.Action),
		zap.String("checkRun", reqData.WorkflowJob.CheckRunURL),
	)

	if reqData.Action == "completed" {
		fmt.Printf("%+v\n\n", reqData)
		util.SendDeployPrompt(endpoints.DiscordSession, endpoints.ChannelID, "neo", "http://cathaloc.dev")
	}

	w.WriteHeader(http.StatusOK)
	response := struct {
		Data string `json:"data"`
	}{
		":)",
	}
	json.NewEncoder(w).Encode(response)

}

func (endpoints *Endpoints) Status(w http.ResponseWriter, r *http.Request) {
	req := datastructures.UptimeResponse{
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
