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

	// Contains these options.
	// 1. path parameter
	// 2. query string
	// 3. request body
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

// func TestConvert_basic(t *testing.T) {
// 	var result OASSchema
// 	mock := `{
// 	"properties" : {
// 		"actionName" : {
// 			"type" : "string",
// 			"description" : "Action Name"
// 		},
// 		"apiDescription" : {
// 			"type" : "string",
// 			"description" : "Api Description"
// 		},
// 		"apiId" : {
// 			"type" : "string",
// 			"description" : "Api Id"
// 		},
// 		"apiName" : {
// 			"type" : "string",
// 			"description" : "Api Name"
// 		},
// 		"disabled" : {
// 			"type" : "boolean",
// 			"description" : "Disabled"
// 		},
// 		"domainCode" : {
// 			"type" : "string",
// 			"description" : "Domain Code"
// 		},
// 		"isDeleted" : {
// 			"type" : "boolean",
// 			"description" : "Is Deleted"
// 		},
// 		"modTime" : {
// 			"type" : "string",
// 			"description" : "Mod Time",
// 			"format" : "date-time"
// 		},
// 		"modifier" : {
// 			"type" : "string",
// 			"description" : "Modifier"
// 		},
// 		"permission" : {
// 			"type" : "string",
// 			"description" : "Permission"
// 		},
// 		"productId" : {
// 			"type" : "string",
// 			"description" : "Product Id"
// 		},
// 		"stages" : {
// 			"type" : "array",
// 			"description" : "Stages",
// 			"items" : {
// 			"type" : "object",
//           "properties" : {
//             "apiId" : {
//               "type" : "string",
//               "description" : "Api Id"
//             },
//             "isPublished" : {
//               "type" : "boolean",
//               "description" : "Is Published"
//             },
//             "stageId" : {
//               "type" : "string",
//               "description" : "Stage Id"
//             },
//             "stageName" : {
//               "type" : "string",
//               "description" : "Stage Name"
//             }
//           }
// 			}
// 		},
// 			"tenantId" : {
// 				"type" : "string",
// 				"description" : "Tenant Id"
// 			}
// 		}
// 	}`

// 	if err := json.Unmarshal([]byte(mock), &result); err != nil {
// 		panic(err)
// 	}

// 	fmt.Println(result)

// 	s1, s2, err := Gen_ConvertOASSchematoTFTypes(result)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(s1, s2)
// }
