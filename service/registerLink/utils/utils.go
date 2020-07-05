package utils

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"regexp"
)

var schemeRe = regexp.MustCompile(`^([a-z][a-zA-Z0-9+\-.]*):`)

func Coalesce(url string) (string, error) {
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

func CreateStatusCodeResponse(statusCode int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
	}
}
