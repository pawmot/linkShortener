package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
	"os"
)

var (
	tableName  = os.Getenv("TABLE_NAME")
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, proxyRequest events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	key := proxyRequest.PathParameters["key"]

	if key == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
		}, nil
	}

	log.Println(fmt.Sprintf("Got a key: %v", key))

	cfg, err := external.LoadDefaultAWSConfig()

	if err != nil {
		log.Println(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	client := dynamodb.New(cfg)

	getResponse, err := client.GetItemRequest(&dynamodb.GetItemInput{
		TableName: &tableName,
		Key: map[string]dynamodb.AttributeValue{
			"linkKey": {S: &key},
		},
		AttributesToGet: []string{
			"url",
		},
	}).Send(ctx)

	if err != nil {
		log.Println(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	url := getResponse.Item["url"]

	log.Println(fmt.Sprintf("Resolved url: %v", url))

	return events.APIGatewayProxyResponse{
		StatusCode: 302,
		Headers: map[string]string{
			"Location": *url.S,
		},
	}, nil
}
