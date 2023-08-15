package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/japb1998/websocket-lambda/internal/connection"
	"github.com/japb1998/websocket-lambda/internal/messages"
)

func handleEvent(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	routeKey := event.RequestContext.RouteKey
	switch routeKey {
	case "$connect":
		{
			var message messages.Event

			message.Type = messages.ENTER_CHAT
			message.Url = fmt.Sprintf("https://%s/%s", event.RequestContext.DomainName, event.RequestContext.Stage)
			message.From = event.Headers["Authorization"]
			message.ConnectionId = event.RequestContext.ConnectionID

			log.Println("$connect", message)

			by, err := json.Marshal(&message)

			if err != nil {
				log.Println(err.Error())
			}

			svc := connection.NewSQSClient()

			input := sqs.SendMessageInput{
				MessageBody: aws.String(string(by)),
				QueueUrl:    aws.String(os.Getenv("QUEUE_URL")),
			}
			_, err = svc.SendMessage(&input)

			if err != nil {
				log.Println(err.Error())
			}

			return events.APIGatewayProxyResponse{
				StatusCode:      200,
				Headers:         nil,
				IsBase64Encoded: false,
			}, nil
		}
	case "$disconnect":
		{
			var message messages.Event

			message.Type = messages.LEAVE_CHAT
			message.Url = fmt.Sprintf("https://%s/%s", event.RequestContext.DomainName, event.RequestContext.Stage)
			message.ConnectionId = event.RequestContext.ConnectionID

			log.Println("$disconnect", message)

			by, err := json.Marshal(&message)

			if err != nil {
				log.Println(err.Error())
			}

			svc := connection.NewSQSClient()

			input := sqs.SendMessageInput{
				MessageBody: aws.String(string(by)),
				QueueUrl:    aws.String(os.Getenv("QUEUE_URL")),
			}
			_, err = svc.SendMessage(&input)

			if err != nil {
				log.Println(err.Error())
			}

			return events.APIGatewayProxyResponse{
				StatusCode:      200,
				Headers:         nil,
				IsBase64Encoded: false,
			}, nil
		}
	case "$default":
		{

			var message messages.Event

			err := json.Unmarshal([]byte(event.Body), &message)

			if err != nil {
				log.Println("$default error:", err.Error())
				return events.APIGatewayProxyResponse{
					StatusCode:      400,
					Headers:         nil,
					IsBase64Encoded: false,
				}, nil
			}

			message.Url = fmt.Sprintf("https://%s/%s", event.RequestContext.DomainName, event.RequestContext.Stage)
			message.From = event.Headers["Authorization"]
			message.ConnectionId = event.RequestContext.ConnectionID

			log.Println("Default", message)

			by, err := json.Marshal(&message)

			if err != nil {
				log.Println(err.Error())
			}

			svc := connection.NewSQSClient()

			input := sqs.SendMessageInput{
				MessageBody: aws.String(string(by)),
				QueueUrl:    aws.String(os.Getenv("QUEUE_URL")),
			}
			_, err = svc.SendMessage(&input)

			if err != nil {
				log.Println(err.Error())
			}

			return events.APIGatewayProxyResponse{
				StatusCode:      200,
				Headers:         nil,
				IsBase64Encoded: false,
			}, nil
		}

	}

	log.Println(event)
	return events.APIGatewayProxyResponse{
		StatusCode:      404,
		Headers:         nil,
		IsBase64Encoded: false,
	}, nil
}
func main() {

	lambda.StartWithOptions(handleEvent)
}
