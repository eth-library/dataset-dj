package main

import (
	"strings"
)

type fileSplit struct {
	localFiles []string
	apiFiles   []string
	cloudFiles []string
}

func splitFiles(files []string) fileSplit {
	localFiles, apiFiles, cloudFiles := []string{}, []string{}, []string{}

	for _, file := range files {
		accessMode, fileName := parseUntil(file, '/')

		if accessMode == "local" {
			localFiles = append(localFiles, fileName)
		} else if accessMode == "api" {
			apiFiles = append(apiFiles, fileName)
		} else {
			cloudFiles = append(cloudFiles, fileName)
		}
	}
	split := fileSplit{
		localFiles: localFiles,
		apiFiles:   apiFiles,
		cloudFiles: cloudFiles,
	}
	return split
}

func parseUntil(file string, delimiter byte) (string, string) {
	index := strings.IndexByte(file, delimiter)
	return file[:index], file[index:]
}
