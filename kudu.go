package td

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type KuduClient struct {
	client        http.Client
	Base          string
	Authorization string
}

func (kudu *KuduClient) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(url, kudu.Base) {
		url = kudu.Base + url
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", kudu.Authorization)
	return req, nil
}

type KuduDeployment struct {
	Id         string `json:"id"`
	Status     int    `json:"status"`
	StatusText string `json:"status_text"`
	Complete   bool   `json:"complete"`
	Active     bool   `json:"active"`
	LogUrl     string `json:"log_url"`
}

// {
//   "id": "ade054e5fa9f475eaa723892b7323f38",
//   "status": 4,
//   "status_text": "",
//   "author_email": "N/A",
//   "author": "N/A",
//   "deployer": "Push-Deployer",
//   "message": "Created via a push deployment",
//   "progress": "",
//   "received_time": "2019-03-21T16:21:13.194086Z",
//   "start_time": "2019-03-21T16:21:14.136643Z",
//   "end_time": "2019-03-21T16:21:19.064429Z",
//   "last_success_end_time": "2019-03-21T16:21:19.064429Z",
//   "complete": true,
//   "active": true,
//   "is_temp": false,
//   "is_readonly": true,
//   "url": "https://tessin-tesla.scm.azurewebsites.net/api/deployments/latest",
//   "log_url": "https://tessin-tesla.scm.azurewebsites.net/api/deployments/latest/log",
//   "site_name": "tessin-tesla",
//   "provisioningState": null
// }

func (kudu *KuduClient) GetDeployment(url string) (*KuduDeployment, error) {
	req, err := kudu.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := kudu.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var deployment KuduDeployment
	err = json.Unmarshal(b, &deployment)
	if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (kudu *KuduClient) Get(url string) ([]byte, error) {
	req, err := kudu.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := kudu.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (kudu *KuduClient) ZipDeploy(filename string) error {
	// see https://github.com/projectkudu/kudu/wiki/Deploying-from-a-zip-file-or-url
	// curl -X POST -u <user> https://{sitename}.scm.azurewebsites.net/api/zipdeploy

	zip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer zip.Close()

	req, err := kudu.NewRequest("POST", "/api/zipdeploy?isAsync=true", zip)
	if err != nil {
		return err
	}

	log.Println("deploying...")

	res, err := kudu.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	log.Println(res.Status)

	deploymentUrl := res.Header.Get("Location")

	for {
		os.Stdout.WriteString(".")

		deployment, err := kudu.GetDeployment(deploymentUrl)
		if err != nil {
			return err
		}

		if deployment.Complete {
			os.Stdout.WriteString("\n")

			b, err := kudu.Get(deployment.LogUrl)
			if err != nil {
				log.Println("could not get deployment log")
				return err
			}
			log.Println(string(b))
			break
		}

		time.Sleep(time.Second)
	}

	return nil
}

type KuduCommand struct {
	Output   string `json:"Output"`
	ExitCode int    `json:"ExitCode"`
}

func (kudu *KuduClient) Command(command string, dir string) (*KuduCommand, error) {
	var x struct {
		Command string `json:"command"`
		Dir     string `json:"dir"`
	}

	x.Command = command
	x.Dir = dir

	b, err := json.Marshal(&x)
	if err != nil {
		return nil, err
	}

	req, err := kudu.NewRequest("POST", "/api/command", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := kudu.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var result KuduCommand
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
