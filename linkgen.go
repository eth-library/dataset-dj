package main

// import (
// 	"fmt"
// 	"time"

// 	"cloud.google.com/go/storage"
// )

// func GenLink(fileName string) string {
// 	method := "GET"
// 	expires := time.Now().Add(time.Hour * 24)

// 	url, err := storage.SignedURL(bucket, fileName, &storage.SignedURLOptions{
// 		GoogleAccessID: accessID,
// 		PrivateKey:     privateKey,
// 		Method:         method,
// 		Expires:        expires,
// 	})
// 	if err != nil {
// 		fmt.Println("error " + err.Error())
// 	}
// 	return url
// }
