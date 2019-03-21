package az

import (
	"io"
	"log"
	"net/http"
)

type AzContext struct {
	client       *http.Client
	ApiVersion   string
	AccessToken  AzAccessToken
	Subscription AzSubscription
}

func GetContext(resource string) (*AzContext, error) {
	accessToken, err := getAccessTokenFromFile(resource)

	if err != nil {
		log.Println("cannot get access token from file")
		return nil, err
	}

	if accessToken == nil {
		log.Println("getting access token from Azure CLI tooling...")

		accessToken, err = getAccessTokenFromAzCli(resource)

		if err != nil {
			return nil, err
		}
	}

	sub, err := getSubscription()
	if err != nil {
		log.Println("cannot figure out subscription, use Azure CLI to set a default")
		return nil, err
	}

	var client http.Client

	return &AzContext{&client, "2018-02-01", *accessToken, *sub}, nil
}

func (ctx *AzContext) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	// todo: refresh access token if expired
	auth := ctx.AccessToken.TokenType + " " + ctx.AccessToken.AccessToken
	req.Header.Add("Authorization", auth)
	return req, nil
}
