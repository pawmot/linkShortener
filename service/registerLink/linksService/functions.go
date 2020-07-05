package linksService

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"os"
	"registerLink/linkKeyCalculator"
	"strconv"
)

var (
	linksTableName   = os.Getenv("LINKS_TABLE_NAME")
	urlIdxName       = os.Getenv("URL_IDX_NAME")
	counterTableName = os.Getenv("COUNTER_TABLE_NAME")
	linkKeyCalc      = linkKeyCalculator.New(314159265359)
)

func ShortenLink(ctx context.Context, url string) (string, error) {
	existingKey, err := checkIfAlreadyShortened(ctx, &url)

	if err != nil {
		return "", err
	}

	var key string
	if existingKey != nil {
		key = *existingKey
		log.Println(fmt.Sprintf("Found an existing key [%s] for the url, will return it", key))
	} else {
		var saved = false

		for !saved {
			cnt, err := getNextCounterValue(ctx)

			if err != nil {
				return "", nil
			}

			key = linkKeyCalc.GetLinkKey(cnt)

			saved, err = tryToSaveShortenedLink(ctx, &key, &url)
		}
	}

	err = memcachedClient.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(url),
		Expiration: 60,
	})

	if err != nil {
		log.Println(fmt.Sprintf("Failed to set cache item: %v", err.Error()))
	} else {
		log.Println("Cache item set successfully")
	}

	return key, nil
}

func checkIfAlreadyShortened(ctx context.Context, url *string) (*string, error) {
	queryResponse, err := ddbClient.QueryRequest(&dynamodb.QueryInput{
		TableName: &linksTableName,
		IndexName: &urlIdxName,
		KeyConditions: map[string]dynamodb.Condition{
			"url": {
				ComparisonOperator: dynamodb.ComparisonOperatorEq,
				AttributeValueList: []dynamodb.AttributeValue{{S: url}},
			},
		},
	}).Send(ctx)

	if err != nil {
		return nil, err
	}

	if *queryResponse.Count > 0 {
		return queryResponse.Items[0]["linkKey"].S, nil
	} else {
		return nil, nil
	}
}

func getNextCounterValue(ctx context.Context) (int, error) {cntIncRes, err := ddbClient.UpdateItemRequest(&dynamodb.UpdateItemInput{
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
		return -1, err
	} else {
		log.Println(fmt.Sprintf("Counter incremented, new counter value: %v", *cntIncRes.Attributes["val"].N))
	}

	return strconv.Atoi(*cntIncRes.Attributes["val"].N)
}

func tryToSaveShortenedLink(ctx context.Context, key *string, url *string) (bool, error) {
	_, err := ddbClient.PutItemRequest(&dynamodb.PutItemInput{
		TableName: &linksTableName,
		Item: map[string]dynamodb.AttributeValue{
			"linkKey": {S: key},
			"url":     {S: url},
		},
		ConditionExpression: aws.String("attribute_not_exists(linkKey)"),
	}).Send(ctx)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.Println(fmt.Sprintf("The linkKey [%s] already exists, continuing to next one", *key))
			default:
				return false, err
			}
		}
	} else {
		log.Println("Link saved to DB")
		return true, nil
	}

	return false, nil
}
