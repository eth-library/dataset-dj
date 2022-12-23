package main

import (
	"fmt"
	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/eth-library/dataset-dj/mailHandler"
	"log"
	"time"
)

func startDownloadLinkEmailTask(url string, recipientEmail string) {
	content := fmt.Sprintf(`
	<h1>Your Download was completed</h1>

	<p>Thanks for using the Data DJ!</p>
	
	<p>
	Please use the link below to retrieve the requested files.
	</p>
	<p>
	   <a href="%v" target="_">%v</a> <br/>
	   (click or copy & paste into your browser)
	</p>
	
	In case of issues, please contact us at contact@librarylab.ethz.ch
	`, url, url)

	emailParts := mailHandler.EmailParts{
		To:         recipientEmail,
		Subject:    "DataDJ - Download complete - Link to retrieve requested files",
		BodyType:   "text/html",
		Body:       content,
		Server:     config.ServiceEmailHost,
		Address:    config.ServiceEmailAddress,
		Password:   config.ServiceEmailPassword,
		ErrorMsg:   "an error occurred while sending the download link email notification: ",
		SuccessMsg: "email with download link sent",
	}
	go mailHandler.SendEmailAsync(emailParts)
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

func inTimeWindow() bool {
	now := getNow()
	return config.StartTime.IsZero() || config.EndTime.IsZero() || inTimeSpan(config.StartTime, config.EndTime, now)
}

func getNow() time.Time {
	nowBase := time.Now()
	now, err := time.Parse("15:04", fmt.Sprintf("%02d:%02d", nowBase.Hour(), nowBase.Minute()))
	if err != nil {
		println(err.Error())
		log.Fatal("Failed to assemble current time defined by layout (15:04)")
	}
	return now
}

func parseTimes(orders []dbutil.Order) []dbutil.OrderTime {
	var res []dbutil.OrderTime
	for _, o := range orders {
		date, _ := time.Parse(time.RFC822, o.Date)
		res = append(res, dbutil.OrderTime{
			OrderID:   o.OrderID,
			ArchiveID: o.ArchiveID,
			Email:     o.Email,
			Date:      date,
			Status:    o.Status,
			Sources:   o.Sources,
		})
	}
	return res
}
