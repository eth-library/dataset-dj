package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eth-library/dataset-dj/dbutil"
	"io"
	"log"
	"net/http"
)

func clientLoop() {
	now := getNow()
	if !config.StartTime.IsZero() && !config.EndTime.IsZero() && !inTimeSpan(config.StartTime, config.EndTime, now) {
		log.Println(fmt.Sprintf("Taskhandler aborting: Outside of designated time interval: (%d:%d - %d:%d)",
			config.StartTime.Hour(), config.StartTime.Minute(), config.EndTime.Hour(),
			config.EndTime.Minute()))
		return
	}

	orders, err := requestOrders()
	if err != nil {
		log.Println(fmt.Sprintf("Taskhandler aborting: %s", err))
		return
	}
	pretty, _ := json.MarshalIndent(orders, "", "  ")
	fmt.Print("orders: \n", string(pretty), "\n")
}

func requestOrders() ([]dbutil.Order, error) {
	url := config.TargetURL + "/handler/orders"
	reqBody, err := json.Marshal(OrderRequestBody{Sources: config.Sources})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal sources to json format: %s", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+config.HandlerKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to request orders from API: %s", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("unable to close response body! Risk of memory leaks if the problem persists")
		}
	}(resp.Body)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %s", err)
	}
	var orderSlice []dbutil.Order
	err = json.Unmarshal(respBody, &orderSlice)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal array of orders: %s", err)
	}
	return orderSlice, nil
}
