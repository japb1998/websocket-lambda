package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dynamodbattribute "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/japb1998/websocket-lambda/internal/client"
	"github.com/japb1998/websocket-lambda/internal/connection"
)

var h = slog.NewTextHandler(os.Stdout, nil).WithAttrs([]slog.Attr{slog.String("pkg", "main")})
var logger = slog.New(h)

func handleEvent(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	routeKey := event.RequestContext.RouteKey
	switch routeKey {
	case "$connect":
		{

			dbclient := client.NewDynamoClient()

			// email is passed from the authorizer, we could techically pass down anything that is needed as long as it serves as a unique identifier
			email := event.RequestContext.Authorizer.(map[string]interface{})["email"].(string)

			record := connection.ConnectionItem{
				User:         email,
				ConnectionId: event.RequestContext.ConnectionID,
			}

			item, err := dynamodbattribute.MarshalMap(record)

			if err != nil {
				slog.Error("error marshalling connection item", slog.String("error", err.Error()))
				return events.APIGatewayProxyResponse{}, fmt.Errorf("error marshalling connection item: %w", err)
			}

			c, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			input := &dynamodb.PutItemInput{
				TableName: aws.String(os.Getenv("CONNECTION_TABLE")),
				Item:      item,
			}

			_, err = dbclient.PutItem(c, input)

			if err != nil {
				logger.Error("error while storing connection id", slog.String("connectionId", event.RequestContext.ConnectionID))
				return events.APIGatewayProxyResponse{}, fmt.Errorf("error while storing connection id: %w", err)
			}

			logger.Info("successfully stored connection id", slog.String("connectionId", event.RequestContext.ConnectionID))

			// store connection id
			return events.APIGatewayProxyResponse{
				StatusCode:      200,
				Headers:         nil,
				IsBase64Encoded: false,
			}, nil
		}
	case "$disconnect":
		{

			dbclient := client.NewDynamoClient()

			key, err := attributevalue.MarshalMap(map[string]string{
				"partitionKey": event.RequestContext.Authorizer.(map[string]interface{})["email"].(string),
				"sortKey":      event.RequestContext.ConnectionID,
			})

			if err != nil {
				logger.Error("error while marshalling key")
				return events.APIGatewayProxyResponse{}, fmt.Errorf("error while marshalling key: %w", err)
			}

			input := &dynamodb.DeleteItemInput{
				TableName: aws.String(os.Getenv("CONNECTION_TABLE")),
				Key:       key,
			}

			c, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			_, err = dbclient.DeleteItem(c, input)

			if err != nil {
				logger.Error("error while deleting connection", slog.Any("error", err.Error()))
				return events.APIGatewayProxyResponse{}, fmt.Errorf("error while deleting connection: %w", err)
			}
			return events.APIGatewayProxyResponse{
				StatusCode:      200,
				Headers:         nil,
				IsBase64Encoded: false,
			}, nil
		}
	case "$default":
		{

			return events.APIGatewayProxyResponse{
				StatusCode:      200,
				Headers:         nil,
				IsBase64Encoded: false,
			}, nil
		}

	}

	return events.APIGatewayProxyResponse{
		StatusCode:      404,
		Headers:         nil,
		IsBase64Encoded: false,
	}, nil
}
func main() {

	lambda.StartWithOptions(handleEvent)
}
