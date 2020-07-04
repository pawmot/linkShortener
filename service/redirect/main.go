package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"os"
)

var (
	tableName     = os.Getenv("TABLE_NAME")
	memcachedAddr = os.Getenv("MEMCACHED_ADDR")
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
	cacheClient := memcache.New(memcachedAddr)

	var url string
	cacheItem, err := cacheClient.Get(key)
	if err == nil {
		url = string(cacheItem.Value)
		log.Println(fmt.Sprintf("Cache hit, resolved url: %v", url))
	} else {
		if err == memcache.ErrCacheMiss {
			log.Println("Cache miss")
		} else {
			log.Println(fmt.Sprintf("Error when hitting cache: %v", err.Error()))
		}

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

		url = *getResponse.Item["url"].S

		log.Println(fmt.Sprintf("Resolved url: %v", url))

		err = cacheClient.Set(&memcache.Item{
			Key:        key,
			Value:      []byte(url),
			Expiration: 60,
		})

		if err != nil {
			log.Println(fmt.Sprintf("Failed to set cache item: %v", err.Error()))
		} else {
			log.Println("Cache item set successfully")
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 302,
		Headers: map[string]string{
			"Location": url,
		},
	}, nil
}
