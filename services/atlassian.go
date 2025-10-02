package services

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ctreminiom/go-atlassian/confluence"
	"github.com/pkg/errors"
)

var (
	clientOnce sync.Once
	client     *confluence.Client
	clientErr  error
)

func loadAtlassianCredentials() (host, mail, token string, err error) {
	host = os.Getenv("ATLASSIAN_HOST")
	mail = os.Getenv("ATLASSIAN_EMAIL")
	token = os.Getenv("ATLASSIAN_TOKEN")

	if host == "" || mail == "" || token == "" {
		return "", "", "", fmt.Errorf("ATLASSIAN_HOST, ATLASSIAN_EMAIL, ATLASSIAN_TOKEN are required environment variables")
	}

	return host, mail, token, nil
}

func ConfluenceClient() (*confluence.Client, error) {
	clientOnce.Do(func() {
		host, mail, token, err := loadAtlassianCredentials()
		if err != nil {
			clientErr = err
			log.Printf("Failed to load Atlassian credentials: %v", err)
			return
		}

		instance, err := confluence.New(nil, host)
		if err != nil {
			clientErr = errors.WithMessage(err, "failed to create confluence client")
			log.Printf("Failed to create Confluence client: %v", clientErr)
			return
		}

		instance.Auth.SetBasicAuth(mail, token)
		client = instance
	})

	if clientErr != nil {
		return nil, clientErr
	}
	if client == nil {
		return nil, fmt.Errorf("confluence client is not initialized")
	}

	return client, nil
}

