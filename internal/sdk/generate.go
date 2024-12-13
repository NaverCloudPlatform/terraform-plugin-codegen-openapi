package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type NClient struct {
	BaseURL    string
	HTTPClient *http.Client
	ACCESS_KEY string
	SECRET_KEY string
}

func NewClient(baseURL, accessKey, secretKey string) *NClient {
	return &NClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
		ACCESS_KEY: accessKey,
		SECRET_KEY: secretKey,
	}
}

type GetProductsProductidApisInfosRequest struct {
	Productid                            string `json:"product-id"`
	ApiName                              string `json:"apiName"`
	HasStage                             string `json:"hasStage"`
	HasStageNotAssociatedWithUsagePlanId string `json:"hasStageNotAssociatedWithUsagePlanId"`
	Limit                                string `json:"limit"`
	Offset                               string `json:"offset"`
	WithStage                            string `json:"withStage"`
}

func (n *NClient) GetProductsProductidApisInfos(r *GetProductsProductidApisInfosRequest) (map[string]interface{}, error) {
	query := map[string]string{
		"apiName":                              r.ApiName,
		"hasStage":                             r.HasStage,
		"hasStageNotAssociatedWithUsagePlanId": r.HasStageNotAssociatedWithUsagePlanId,
		"limit":                                r.Limit,
		"offset":                               r.Offset,
		"withStage":                            r.WithStage,
	}

	body := ""

	url := n.BaseURL + "/" + "products" + "/" + r.Productid + "/" + "apis" + "/" + "infos"

	response, err := n.MakeRequest("GET", url, body, query)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("output is nil")
	}

	return response, nil
}

type PostProductsProductidApisCloneRequest struct {
	Productid      string `json:"product-id"`
	ApiDescription string `json:"apiDescription"`
	ApiName        string `json:"apiName"`
	OriginApiId    string `json:"originApiId"`
}

func (n *NClient) PostProductsProductidApisClone(r *PostProductsProductidApisCloneRequest) (map[string]interface{}, error) {
	query := map[string]string{}

	rawBody, err := json.Marshal(map[string]string{
		"apiDescription": r.ApiDescription,
		"apiName":        r.ApiName,
		"originApiId":    r.OriginApiId,
	})
	if err != nil {
		return nil, err
	}

	body := strings.Replace(string(rawBody), `\"`, "", -1)

	url := n.BaseURL + "/" + "products" + "/" + r.Productid + "/" + "apis" + "/" + "clone"

	response, err := n.MakeRequest("POST", url, body, query)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("output is nil")
	}

	return response, nil
}
