package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/IamCathal/neo/services/frontend/configuration"
	"github.com/neosteamfriendgraphing/common"
)

type Cntr struct{}

type CntrInterface interface {
	SaveCrawlingStats(crawlingStatusJSON []byte) (bool, error)
}

func (control Cntr) SaveCrawlingStats(crawlingStatusJSON []byte) (bool, error) {
	targetURL := fmt.Sprintf("%s/savecrawlingstats", os.Getenv("DATASTORE_URL"))
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(crawlingStatusJSON))
	if err != nil {
		configuration.Logger.Sugar().Infof("error creating POST /savecrawlingstats request: %+v")
		return false, err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", "something")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		configuration.Logger.Sugar().Infof("error calling /savecrawlingstats: %+v", err)
		return false, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		configuration.Logger.Sugar().Infof("error reading body from /savecrawlingstats: %+v", err)
		return false, err
	}
	APIRes := common.BasicAPIResponse{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		configuration.Logger.Sugar().Infof("error unmarshaling body from /savecrawlingstats: %+v", err)
		return false, err
	}

	if APIRes.Status == "success" && res.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, fmt.Errorf("failed to save crawling status: %+v", APIRes)
}
