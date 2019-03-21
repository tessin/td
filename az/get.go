package az

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/tessin/td/jwt"
)

type AzAccessToken struct {
	TokenType   string `json:"tokenType"`
	AccessToken string `json:"accessToken"`
}

const (
	ResourceManager = "https://management.core.windows.net/"
)

func getAzureAzCliPath(fn string) (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHomeDir, ".azure", fn), nil
}

// get the access token from the file .azure/accessTokens.json cache
func getAccessTokenFromFile(resource string) (*AzAccessToken, error) {
	accessTokensJsonFn, err := getAzureAzCliPath("accessTokens.json")
	if err != nil {
		return nil, err
	}

	b, err := readTextFile(accessTokensJsonFn)
	if err != nil {
		return nil, err
	}

	var accessTokens []AzAccessToken
	err = json.Unmarshal(b, &accessTokens)
	if err != nil {
		return nil, err
	}

	for _, accessToken := range accessTokens {
		claims, err := jwt.Decode(accessToken.AccessToken)
		if err != nil {
			log.Println("cannot decode access token", err)
			continue
		}

		if claims.Aud() == resource {
			if claims.Exp().After(time.Now()) {
				log.Println("using token from cache, will expire in", claims.Exp().Sub(time.Now()))
				return &accessToken, nil
			}
		}
	}

	return nil, nil
}

type AzSubscription struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

type AzProfile struct {
	Subscriptions []AzSubscription `json:"subscriptions"`
}

func readTextFile(fn string) ([]byte, error) {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	// UTF-8 byte order mark
	if bytes.HasPrefix(b, []byte{0xef, 0xbb, 0xbf}) {
		return b[3:], nil
	}

	return b, nil
}

func getSubscription() (*AzSubscription, error) {
	azureProfileJsonFn, err := getAzureAzCliPath("azureProfile.json")
	if err != nil {
		return nil, err
	}

	b, err := readTextFile(azureProfileJsonFn)
	if err != nil {
		return nil, err
	}

	var profile AzProfile
	err = json.Unmarshal(b, &profile)
	if err != nil {
		return nil, err
	}

	for _, sub := range profile.Subscriptions {
		if sub.IsDefault {
			return &sub, nil
		}
	}

	return nil, nil
}

func getAccessTokenFromAzCli(resource string) (*AzAccessToken, error) {
	var out bytes.Buffer

	cmd := exec.Command("az", "account", "get-access-token", "--resource", resource)
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var token AzAccessToken
	err = json.Unmarshal(out.Bytes(), &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}
