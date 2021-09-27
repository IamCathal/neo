package main

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
	"strings"
	"time"

	"github.com/IamCathal/neo/services/pluto/datastructures"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (endpoints *Endpoints) HomeHandler(w http.ResponseWriter, req *http.Request) {
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// fmt.Fprintf(w, "working")
	// endpoints.logger.Info("called / endpoint")
	panic("blah blah")
}

func (endpoints *Endpoints) status(w http.ResponseWriter, r *http.Request) {
	time.Sleep(400 * time.Millisecond)

	req := datastructures.UptimeResponse{
		Uptime: time.Since(applicationStartUpTime),
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

func (endpoints *Endpoints) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	gotHash := strings.SplitN(r.Header.Get("X-Hub-Signature-256"), "=", 2)
	if gotHash[0] != "sha256" {
		endpoints.logger.Error(fmt.Sprintf("invalid format signature '%s' received", gotHash[0]),
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", http.StatusInternalServerError),
			zap.String("path", r.URL.EscapedPath()),
		)
		panic("invalid signature format received")
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		endpoints.logger.Error(fmt.Sprintf("Cannot read the request body: %s\n", err),
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", http.StatusInternalServerError),
			zap.String("path", r.URL.EscapedPath()),
		)
		panic("cannot read request body")
	}
	hash := hmac.New(sha256.New, []byte(os.Getenv("WEBHOOK_SECRET")))
	if _, err := hash.Write(b); err != nil {
		endpoints.logger.Error(fmt.Sprintf("Cannot compute the HMAC for request: %s\n", err),
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", http.StatusInternalServerError),
			zap.String("path", r.URL.EscapedPath()),
		)
		panic("cannot compute HMAC")
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))
	if gotHash[1] != expectedHash {
		endpoints.logger.Error(fmt.Sprintf("invalid signature '%s' does not match expected '%s'", gotHash[1], expectedHash),
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
		endpoints.logger.Error(fmt.Sprintf("could not unmarshal request body: %s", err),
			zap.String("requestID", vars["requestID"]),
			zap.Int("status", http.StatusInternalServerError),
			zap.String("path", r.URL.EscapedPath()),
		)
		return
	}

	endpoints.logger.Info(fmt.Sprintf("%s was run with state: %s", reqData.WorkflowJob.Name, reqData.Action),
		zap.String("checkRun", reqData.WorkflowJob.CheckRunURL),
	)

	if reqData.Action == "completed" {
		fmt.Printf("%+v\n\n", reqData)
		SendDeployPrompt(DiscordSession, "neo", "http://cathaloc.dev")
	}

	w.WriteHeader(http.StatusOK)
	response := struct {
		Data string `json:"data"`
	}{
		":)",
	}
	json.NewEncoder(w).Encode(response)

}
