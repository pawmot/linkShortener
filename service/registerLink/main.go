package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"os"
	"regexp"
	"registerLink/linkKeyCalculator"
	"strconv"
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
	linksTableName   = os.Getenv("LINKS_TABLE_NAME")
	urlIdxName       = os.Getenv("URL_IDX_NAME")
	counterTableName = os.Getenv("COUNTER_TABLE_NAME")
	memcachedAddr    = os.Getenv("MEMCACHED_ADDR")
	linkKeyProvider  = linkKeyCalculator.New(314159265359)
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

	log.Println(fmt.Sprintf("Got url: %v", req.Url))

	url, err := coalesce(req.Url)

	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, nil
	}

	if url != req.Url {
		log.Println(fmt.Sprintf("Shortening coalesced url: %v", url))
	}

	cfg, err := external.LoadDefaultAWSConfig()

	if err != nil {
		log.Println(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	dynamoClient := dynamodb.New(cfg)
	cacheClient := memcache.New(memcachedAddr)

	queryResponse, err := dynamoClient.QueryRequest(&dynamodb.QueryInput{
		TableName: &linksTableName,
		IndexName: &urlIdxName,
		KeyConditions: map[string]dynamodb.Condition{
			"url": {
				ComparisonOperator: dynamodb.ComparisonOperatorEq,
				AttributeValueList: []dynamodb.AttributeValue{{S: &url}},
			},
		},
	}).Send(ctx)

	if err != nil {
		log.Println(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	var key string
	if *queryResponse.Count > 0 {
		key = *queryResponse.Items[0]["linkKey"].S
		log.Println(fmt.Sprintf("Found an existing key [%s] for the url, will return it", key))
	} else {
		var saved = false

		for !saved {
			cntIncRes, err := dynamoClient.UpdateItemRequest(&dynamodb.UpdateItemInput{
				TableName: &counterTableName,
				Key: map[string]dynamodb.AttributeValue{
					"cnt_name": {S: aws.String("linkKeyCnt")},
				},
				UpdateExpression: aws.String("SET val = val + :inc"),
				ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
					":inc": {N: aws.String("1")},
				},
				ReturnValues: dynamodb.ReturnValueUpdatedNew,
			}).Send(ctx)

			if err != nil {
				log.Println(err.Error())
				return events.APIGatewayProxyResponse{
					StatusCode: 500,
				}, nil
			} else {
				log.Println(fmt.Sprintf("Counter incremented, new counter value: %v", *cntIncRes.Attributes["val"].N))
			}

			cnt, err := strconv.Atoi(*cntIncRes.Attributes["val"].N)

			if err != nil {
				log.Println("Cannot parse counter value")
				return events.APIGatewayProxyResponse{
					StatusCode: 500,
				}, nil
			}

			key = linkKeyProvider.GetLinkKey(cnt)

			_, err = dynamoClient.PutItemRequest(&dynamodb.PutItemInput{
				TableName: &linksTableName,
				Item: map[string]dynamodb.AttributeValue{
					"linkKey": {S: &key},
					"url":     {S: &url},
				},
				ConditionExpression: aws.String("attribute_not_exists(linkKey)"),
			}).Send(ctx)

			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case dynamodb.ErrCodeConditionalCheckFailedException:
						log.Println(fmt.Sprintf("The linkKey [%s] already exists, continuing to next one", key))
					default:
						log.Println(err.Error())
						return events.APIGatewayProxyResponse{
							StatusCode: 500,
						}, nil
					}
				}
			} else {
				log.Println("Link saved to DB")
				saved = true
			}
		}
	}

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

	res := RegisterLinkResponse{
		ShortenedLink: fmt.Sprintf("https://1cpmx32uyj.execute-api.eu-west-1.amazonaws.com/%v", key),
	}

	resStr, err := json.Marshal(res)

	if err != nil {
		log.Println(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(resStr),
	}, nil
}

var schemeRe = regexp.MustCompile(`^([a-z][a-zA-Z0-9+\-.]*):`)

func coalesce(url string) (string, error) {
	match := schemeRe.FindStringSubmatch(url)
	if len(match) == 2 {
		scheme := match[1]
		if scheme == "http" || scheme == "https" {
			return url, nil
		} else {
			return "", fmt.Errorf("unhandled scheme: %s", scheme)
		}
	} else {
		return "http://" + url, nil
	}
}
