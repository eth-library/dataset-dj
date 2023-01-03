package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eth-library/dataset-dj/constants"
	"github.com/eth-library/dataset-dj/dbutil"
	"io"
	"log"
	"net/http"
	"os"
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
			var order dbutil.TimedOrder
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

func fulfillOrder(order dbutil.TimedOrder) error {
	err := acknowledgeOrder(order.OrderID, constants.Processing)
	if err != nil {
		return err
	}
	// ----------------------------
	// This is where the files are downloaded, compressed and stored in a new location
	url, err := zipFiles(order)
	time.Sleep(15 * time.Second) // Only for testing purposes, should be removed afterwards
	// ----------------------------
	err = acknowledgeOrder(order.OrderID, constants.Closed)
	fmt.Printf("Order %s has been fulfilled and the files from archive %s were downloaded\n",
		order.OrderID, order.ArchiveID)
	if err != nil {
		return err
	}
	startDownloadLinkEmailTask(url, order.Email)
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

func requestOrders() ([]dbutil.TimedOrder, error) {
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

func requestArchive(archiveID string) (*dbutil.MetaArchive, error) {
	url := config.TargetURL + "/archive/" + archiveID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+config.HandlerKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to request archive from API: %s", err)
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
	var archive dbutil.MetaArchive
	err = json.Unmarshal(respBody, &archive)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal meta archive: %s", err)
	}
	return &archive, nil
}

func zipFiles(order dbutil.TimedOrder) (string, error) {
	// fmt.Println("creating local zip archive...")
	archiveFilePath := config.ArchiveDir + "/" + constants.ArchiveBaseName + "_" + order.ArchiveID + ".zip"
	zipPath, err := os.Create(archiveFilePath)
	if err != nil {
		log.Print("ERROR: while creating local zip file:", err)
	}
	defer zipPath.Close()
	zipWriter := zip.NewWriter(zipPath)
	archive, err := requestArchive(order.ArchiveID)
	if err != nil {
		return "", fmt.Errorf("error while retrieving archive information: %s", err)
	}

	for _, fg := range archive.Content {
		for i, file := range fg.Files.ToSlice() {
			err := WriteLocalToZip(file, zipWriter)
			if err != nil {
				fmt.Printf("\r zipping file %d / %d: %s\n", i+1, len(fg.Files.ToSlice()), file)
				log.Print(err)
			}
		}
	}

	err = zipWriter.Close()
	if err != nil {
		return "", fmt.Errorf("unable to close response body!"+
			" Risk of memory leaks if the problem persists: %s", err)
	}
	return archiveFilePath, nil
}
