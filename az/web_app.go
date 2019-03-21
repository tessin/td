package az

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

type Resource struct {
	Id         string          `json:"id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Location   string          `json:"location"`
	Properties json.RawMessage `json:"properties"`
}

func (r Resource) String() string {
	return r.Name
}

type ResourceList struct {
	Value []Resource `json:"value"`
}

func (ctx *AzContext) GetWebSiteList(resourceGroup string) (*ResourceList, error) {
	url := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%v/resourceGroups/%v/providers/Microsoft.Web/sites?api-version="+ctx.ApiVersion,
		ctx.Subscription.Id,
		resourceGroup,
	)

	req, err := ctx.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := ctx.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var list ResourceList
	err = json.Unmarshal(b, &list)
	if err != nil {
		return nil, err
	}

	return &list, nil
}

type WebPublishData struct {
	PublishProfiles []WebPublishProfile `xml:"publishProfile"`
}

type WebPublishProfile struct {
	User string `xml:"userName,attr"`
	Pass string `xml:"userPWD,attr"`
}

func (ctx *AzContext) GetWebPublishXml(resourceGroup, name string) (*WebPublishProfile, error) {
	// https://management.azure.com/subscriptions/1c4992c2-c303-4972-8122-8d3a28201cd9/resourceGroups/TCM/providers/Microsoft.Web/sites/tessin-tesla/publishxml?api-version=2018-02-01

	url := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%v/resourceGroups/%v/providers/Microsoft.Web/sites/%v/publishxml?api-version="+ctx.ApiVersion,
		ctx.Subscription.Id,
		resourceGroup,
		name,
	)

	req, err := ctx.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := ctx.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var publishData WebPublishData
	err = xml.Unmarshal(b, &publishData)
	if err != nil {
		return nil, err
	}

	return &publishData.PublishProfiles[0], nil
}
