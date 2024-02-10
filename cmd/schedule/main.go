package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	dynamodbattribute "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/japb1998/websocket-lambda/internal/client"
	"github.com/japb1998/websocket-lambda/internal/connection"
)

var h = slog.NewTextHandler(os.Stdout, nil).WithAttrs([]slog.Attr{slog.String("pkg", "schedule")})
var logger = slog.New(h)

func handler(ctx context.Context, e interface{}) error {

	var wsMessage struct {
		Body      string `json:"body"`
		CreatedAt string `json:"createdAt"`
	} = struct {
		Body      string "json:\"body\""
		CreatedAt string "json:\"createdAt\""
	}{
		Body:      "Medium Websocket Message",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	url := os.Getenv("WS_URL")

	apiGtw := client.NewApiGatewayClient(url)

	data, err := json.Marshal(wsMessage)

	if err != nil {
		logger.Error("error marshalling message", slog.String("error", err.Error()))
		return err
	}
	var lastEvaluatedKey map[string]types.AttributeValue
	items := make([]connection.ConnectionItem, 0)

	// retrieve all the connection ids from the dynamoDB table
	for {
		dbclient := client.NewDynamoClient()

		scanInput := &dynamodb.ScanInput{
			TableName:         aws.String(os.Getenv("CONNECTION_TABLE")),
			ExclusiveStartKey: lastEvaluatedKey,
		}

		c, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		output, err := dbclient.Scan(c, scanInput)

		if err != nil {
			return err
		}

		var dynamoRecords []connection.ConnectionItem

		err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &dynamoRecords)

		items = append(items, dynamoRecords...)

		if output.LastEvaluatedKey == nil {
			break
		}
	}

	var wg sync.WaitGroup

	for _, dynamoRecord := range items {
		wg.Add(1)

		// we do  not have to worry about the loop variable since we are using golang 1.22
		go func() {
			defer wg.Done()

			c, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			_, err = apiGtw.PostToConnection(c, &apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: &dynamoRecord.ConnectionId,
				Data:         data,
			})

			if err != nil {
				logger.Error("error while broadcasting chat", slog.String("error", err.Error()))
			}
		}()
	}

	wg.Wait()

	return nil
}
func main() {

	lambda.Start(handler)
}
