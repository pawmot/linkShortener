package linksService

import (
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/bradfitz/gomemcache/memcache"
	"os"
)

var ddbClient = getDynamodbClient()
var memcachedClient = getMemcachedClient()

var memcachedAddr = os.Getenv("MEMCACHED_ADDR")

func getDynamodbClient() *dynamodb.Client {

	cfg, err := external.LoadDefaultAWSConfig()

	if err != nil {
		panic(err)
	}

	return dynamodb.New(cfg)
}

func getMemcachedClient() *memcache.Client {
	return memcache.New(memcachedAddr)
}
