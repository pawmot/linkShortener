package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
	"os"
)

func main() {
	lambda.Start(HandleRequest)
}

type RegisterLinkRequest struct {
	Url string `json:"url"`
}

type RegisterLinkResponse struct {
	ShortenedLink string `json:"shortenedLink"`
}

var (
	tableName  = os.Getenv("TABLE_NAME")
	urlIdxName = os.Getenv("URL_IDX_NAME")
)

func HandleRequest(ctx context.Context, proxyRequest events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req := RegisterLinkRequest{}
	err := json.Unmarshal([]byte(proxyRequest.Body), &req)

	if err != nil {
		log.Println(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
		}, nil
	}

	log.Println(fmt.Sprintf("Shortening url: %v", req.Url))

	cfg, err := external.LoadDefaultAWSConfig()

	if err != nil {
		log.Println(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	client := dynamodb.New(cfg)

	hash := fmt.Sprintf("%x", md5.Sum([]byte(req.Url)))

	_, err = client.PutItemRequest(&dynamodb.PutItemInput{
		TableName: &tableName,
		Item: map[string]dynamodb.AttributeValue{
			"linkKey": {S: &hash},
			"url":     {S: &req.Url},
		},
	}).Send(ctx)

	if err != nil {
		log.Println(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	res := RegisterLinkResponse{
		ShortenedLink: fmt.Sprintf("https://1cpmx32uyj.execute-api.eu-west-1.amazonaws.com/%v", hash),
	}

	resStr, err := json.Marshal(res)

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(resStr),
	}, nil
}
