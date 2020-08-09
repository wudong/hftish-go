package logging

import (
	gclogging "cloud.google.com/go/logging"
	"context"
	"log"
	"os"
	"strings"
)

var client *gclogging.Client
var logger *log.Logger

func GetLogger() *log.Logger {
	if nil != logger {
		return logger
	} else {
		localDev := os.Getenv("DEV_LOGGING")
		if strings.ToLower(localDev) == "true" {
			return log.New(os.Stdout, "main", log.LstdFlags)
		} else { //using google stack drive.
			return newCloudLogger()
		}
	}
}

func newCloudLogger() *log.Logger {
	ctx := context.Background()

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		panic("Need to specify GCP_PROJECT_ID environment variable for using gcp stackdriver logging")
	}

	lClient, err := gclogging.NewClient(ctx, projectID)
	if err != nil {
		panic("Error creating client to cloud logger")
	}
	client = lClient
	logger := client.Logger("hftish")
	return logger.StandardLogger(gclogging.Info)
}

func Cleanup() {
	if client != nil {
		client.Close()
		client = nil
	}
}
