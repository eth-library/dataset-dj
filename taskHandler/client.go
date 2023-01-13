package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/eth-library/dataset-dj/constants"
	"github.com/eth-library/dataset-dj/dbutil"
)

func clientLoop() {
	if !inTimeWindow() {
		log.Printf("Taskhandler aborting: Outside of designated time interval: (%d:%d - %d:%d)",
			config.StartTime.Hour(), config.StartTime.Minute(), config.EndTime.Hour(),
			config.EndTime.Minute())
		return
	}

	orders, err := requestOrders()
	if err != nil {
		log.Printf("Taskhandler aborting: %s", err)
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
			log.Printf("Taskhandler aborting: Outside of designated time interval: (%d:%d - %d:%d)",
				config.StartTime.Hour(), config.StartTime.Minute(), config.EndTime.Hour(),
				config.EndTime.Minute())
			return
		}
		if len(orders) == 0 {
			orders, err = requestOrders()
			if err != nil {
				log.Printf("Taskhandler aborting: %s", err)
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

	url, err := fetch_n_zipFiles(order)
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second) // Only for testing purposes, should be removed afterwards
	// ----------------------------
	// libDrive gezauber
	err = push2libDrive(order)
	if err != nil {
		return err
	}
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
	if err != nil {
		fmt.Println("unable to build httpRequest: ", err)
	}
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
	if err != nil {
		fmt.Println("unable to build httpRequest: ", err)
	}
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

func requestArchive(archiveID string) (*Archive, error) {
	fmt.Println("requesting archive from api...")
	url := config.TargetURL + "/handler/archive/" + archiveID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("unable to build httpRequest: ", err)
	}
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
	if strings.Contains(string(respBody), "Insufficient Token Permission for Request") {
		return nil, fmt.Errorf("error: %s", string(respBody))
	}

	var archive Archive
	err = json.Unmarshal(respBody, &archive)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal meta archive: %s", err)
	}
	// fmt.Println(archive.ID, " : ", archive.Content)
	return &archive, nil
}

func fetch_n_zipFiles(order dbutil.TimedOrder) (string, error) {
	fmt.Println("creating local zip archive...")
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
	// TODO: hier muss eine Lösung rein, an die Files zu kommen. Erster Workaround nur für ePeriodica: der Server wird hier rein gemountet
	for _, fg := range archive.Content {
		// add meta-json
		err := os.WriteFile("meta.json", []byte(archive.Meta), 0644)
		if err != nil {
			fmt.Println("error writing meta.json: ", err)
		}
		fg.Files = append(fg.Files, "meta.json")
		for i, filename := range fg.Files {
			err := WriteLocalToZip(filename, zipWriter)
			if err != nil {
				fmt.Printf("\r zipping file %d / %d: %s\n", i+1, len(fg.Files), filename)
				log.Print(err)
			}
		}
		err = os.Remove("meta.json")
		if err != nil {
			fmt.Println("error removing meta.json: ", err)
		}
	}

	err = zipWriter.Close()
	if err != nil {
		return "", fmt.Errorf("unable to close response body!"+
			" Risk of memory leaks if the problem persists: %s", err)
	}
	return archiveFilePath, nil
}

func push2libDrive(order dbutil.TimedOrder) error {
	err := createShare(order)
	if err != nil {
		return err
	}

	return nil
}

func createShare(order dbutil.TimedOrder) error {
	base_uri := libDriveConfig.Host + libDriveConfig.ApiPath
	username := secrets.LibDriveUser
	password := secrets.LibDrivePassword
	ethzuser := strings.Split(order.Email, "@")[0]
	fedshareaddress := ethzuser + "@polybox.ethz.ch"
	archiveFileName := constants.ArchiveBaseName + "_" + order.ArchiveID + ".zip"

	delivery_url := libDriveConfig.Host + libDriveConfig.WebdavPath + username + "/orders/" + order.OrderID
	client := &http.Client{}
	// 1. create folder
	// curl -u $USER:$PASSWD -X MKCOL \
	//	"$HOST$WEBDAVPATH/$FOLDER"
	req, err := http.NewRequest("MKCOL", delivery_url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(username, password)
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	// 2. push Archive to folder
	// curl -u $USER:$PASSWD -T $FILE \
	//   "$HOST$WEBDAVPATH/$FOLDER/$FILE"
	data, err := os.Open(config.ArchiveDir + "/" + archiveFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer data.Close()
	req, err = http.NewRequest("PUT", delivery_url+"/"+archiveFileName, data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/zip")
	req.SetBasicAuth(username, password)
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	// 3. create share
	// Set the form POST body
	form := url.Values{}
	form.Add("path", "/orders/"+order.OrderID)
	form.Add("shareType", "6") // federated share
	form.Add("shareWith", fedshareaddress)
	form.Add("permissions", "31")

	// Build the core request object
	req, _ = http.NewRequest(
		"POST",
		fmt.Sprintf("%s/%s", base_uri, "shares"),
		strings.NewReader(form.Encode()),
	)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(bodyText))

	return nil
}
