package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IamCathal/neo/services/pluto/endpoints"
	"github.com/IamCathal/neo/services/pluto/util"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func startWebServer(r *mux.Router) {
	log.Printf("Starting web server on http://%s:%s\n", util.GetLocalIPAddress(), os.Getenv("API_PORT"))
	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("API_PORT")), r)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	logConfig, err := util.LoadLoggingConfig()
	if err != nil {
		log.Fatal(err)
	}
	logger := util.InitLogger(logConfig)

	endpoints := &endpoints.Endpoints{
		Logger:                 logger,
		ApplicationStartUpTime: time.Now(),
		ChannelID:              os.Getenv("DISCORD_CHANNEL_ID"),
	}

	// Start web server
	r := mux.NewRouter()
	r.HandleFunc("/status", endpoints.Status).Methods("POST")
	r.Use(endpoints.LoggingMiddleware)
	authEndpoints := r.PathPrefix("/api").Subrouter()
	authEndpoints.HandleFunc("/input", endpoints.ReceiveWebhook).Methods("POST")
	go startWebServer(r)

	// Start the discord bot
	Session, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		logger.Fatal(fmt.Sprintf("error creating discord session: %s", err))
		return
	}
	Session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
	err = Session.Open()
	if err != nil {
		logger.Fatal(fmt.Sprintf("error opening discord session: %s", err))
		return
	}
	defer Session.Close()

	endpoints.DiscordSession = Session

	logger.Info("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
