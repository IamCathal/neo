package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

var (
	applicationStartUpTime time.Time

	nodeName string
	nodeDC   string
	logPath  string
	nodeIPV4 string

	// DiscordBot variables
	ChanneID       string
	DiscordSession *discordgo.Session
)

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

type Endpoints struct {
	logger *zap.Logger
}

func (endpoints *Endpoints) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("the err %s\n\n", err)
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
					endpoints.logger.Fatal(fmt.Sprintf("req start time err: %v", err),
						zap.String("requestID", vars["requestID"]),
						zap.Int("status", http.StatusInternalServerError),
						// zap.Int64("duration", GetCurrentTimeInMs()-requestStartTime),
						zap.String("path", r.URL.EscapedPath()),
					)
					panic(timeParseErr)
				}

				endpoints.logger.Error(fmt.Sprintf("normal err: %v", err),
					zap.String("requestID", vars["requestID"]),
					zap.Int("status", http.StatusInternalServerError),
					zap.Int64("duration", GetCurrentTimeInMs()-requestStartTime),
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

		endpoints.logger.Info("hello world",
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", wrapped.status),
			zap.Int64("duration", GetCurrentTimeInMs()-requestStartTime),
			zap.String("path", r.URL.EscapedPath()),
		)
	})
}

func SendDeployPrompt(Session *discordgo.Session, serviceName, actionsUrl string) {
	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x2d95bc, // Green
		Description: "Tests passed successfully",
		URL:         actionsUrl,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3b/Eo_circle_green_checkmark.svg/2048px-Eo_circle_green_checkmark.svg.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Title:     fmt.Sprintf("Would you like to deploy %s?", serviceName),
	}
	Session.ChannelMessageSendEmbed(ChanneID, embed)
}

func startWebServer(r *mux.Router) {
	log.Printf("Starting web server on http://%s:%s\n", GetLocalIPAddress(), os.Getenv("API_PORT"))
	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("API_PORT")), r)
}
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	applicationStartUpTime = time.Now()
	LoadConfig()

	ChanneID = os.Getenv("DISCORD_CHANNEL_ID")

	endpoints := &Endpoints{
		logger: InitLogger(),
	}
	r := mux.NewRouter()
	r.HandleFunc("/status", endpoints.status).Methods("POST")
	r.Use(endpoints.LoggingMiddleware)
	authEndpoints := r.PathPrefix("/api").Subrouter()
	authEndpoints.HandleFunc("/input", endpoints.ReceiveWebhook).Methods("POST")

	go startWebServer(r)

	Session, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	Session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
	err = Session.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	defer Session.Close()

	DiscordSession = Session

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}
