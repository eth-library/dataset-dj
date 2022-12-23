package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eth-library/dataset-dj/constants"
	"github.com/eth-library/dataset-dj/dbutil"
	"io"
	"log"
	"net/http"
	"sort"
	"time"
)

func clientLoop() {
	if !inTimeWindow() {
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

	for {
		if len(orders) != 0 {
			var order dbutil.OrderTime
			order, orders = orders[0], orders[1:]
			err := fulfillOrder(order)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			time.Sleep(time.Duration(config.RequestInterval) * time.Second)
		}
		if !inTimeWindow() {
			log.Println(fmt.Sprintf("Taskhandler aborting: Outside of designated time interval: (%d:%d - %d:%d)",
				config.StartTime.Hour(), config.StartTime.Minute(), config.EndTime.Hour(),
				config.EndTime.Minute()))
			return
		}
		if len(orders) == 0 {
			orders, err = requestOrders()
			if err != nil {
				log.Println(fmt.Sprintf("Taskhandler aborting: %s", err))
				return
			}
		}
	}
}

func fulfillOrder(order dbutil.OrderTime) error {
	err := acknowledgeOrder(order.OrderID, constants.Processing)
	if err != nil {
		return err
	}
	// This is where the files are downloaded, compressed and stored in a new location
	time.Sleep(15 * time.Second)
	err = acknowledgeOrder(order.OrderID, constants.Closed)
	fmt.Printf("Order %s has been fulfilled and the files from archive %s were downloaded\n",
		order.OrderID, order.ArchiveID)
	if err != nil {
		return err
	}
	startDownloadLinkEmailTask("dummy-url", order.Email)
	return nil
}

func acknowledgeOrder(orderID string, status string) error {
	url := config.TargetURL + "/handler/order/" + orderID
	reqBody, err := json.Marshal(OrderStatusBody{NewStatus: status})
	if err != nil {
		return fmt.Errorf("unable to marshal status to json format: %s", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+config.HandlerKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to change status of order: %s", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("unable to close response body! Risk of memory leaks if the problem persists")
		}
	}(resp.Body)
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return fmt.Errorf("unable to change status of order")
	}
}

func requestOrders() ([]dbutil.OrderTime, error) {
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
	orders := parseTimes(orderSlice)
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].Date.Before(orders[j].Date)
	})
	return orders, nil
}
