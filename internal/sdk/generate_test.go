package sdk

import (
	"fmt"
	"os"
	"testing"
)

func TestApiInfo_basic(t *testing.T) {

	n := NewClient("https://apigateway.apigw.ntruss.com/api/v1", os.Getenv("NCLOUD_ACCESS_KEY"), os.Getenv("NCLOUD_SECRET_KEY"))

	r := &GetProductsProductidApisInfosRequest{
		Productid: "9aip5utnp2",
		HasStage:  "false",
	}

	res, err := n.GetProductsProductidApisInfos(r)
	if err != nil {
		panic(err)
	}

	fmt.Printf("===================== %+v", res)
}

func TestApiClone_basic(t *testing.T) {

	n := NewClient("https://apigateway.apigw.ntruss.com/api/v1", os.Getenv("NCLOUD_ACCESS_KEY"), os.Getenv("NCLOUD_SECRET_KEY"))

	r := &PostProductsProductidApisCloneRequest{
		Productid:      "9aip5utnp2",
		ApiDescription: "test description",
		ApiName:        "api1clone",
		OriginApiId:    "ia9u5xzfb0",
	}

	res, err := n.PostProductsProductidApisClone(r)
	if err != nil {
		panic(err)
	}

	fmt.Printf("===================== %+v", res)
}
