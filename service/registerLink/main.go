package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"registerLink/linksService"
	"registerLink/utils"
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

func HandleRequest(ctx context.Context, proxyRequest events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req := RegisterLinkRequest{}
	err := json.Unmarshal([]byte(proxyRequest.Body), &req)

	if err != nil {
		log.Println(err.Error())
		return utils.CreateStatusCodeResponse(400), nil
	}

	log.Println(fmt.Sprintf("Got url: %v", req.Url))

	url, err := utils.Coalesce(req.Url)

	if err != nil {
		log.Println(err.Error())
		return utils.CreateStatusCodeResponse(400), nil
	}

	if url != req.Url {
		log.Println(fmt.Sprintf("Coalesced url: %v", url))
	}

	key, err := linksService.ShortenLink(ctx, url)

	if err != nil {
		log.Println(err)
		return utils.CreateStatusCodeResponse(500), nil
	}

	res := RegisterLinkResponse{
		ShortenedLink: fmt.Sprintf("https://1cpmx32uyj.execute-api.eu-west-1.amazonaws.com/%v", key),
	}

	resStr, err := json.Marshal(res)

	if err != nil {
		return utils.CreateStatusCodeResponse(500), nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(resStr),
	}, nil
}
