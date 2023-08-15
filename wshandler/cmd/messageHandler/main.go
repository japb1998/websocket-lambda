package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/japb1998/websocket-lambda/internal/connection"
	"github.com/japb1998/websocket-lambda/internal/messages"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {

	for _, record := range sqsEvent.Records {

		var event messages.Event

		err := json.Unmarshal([]byte(record.Body), &event)

		if err != nil {
			log.Println(err.Error())
			continue
		}

		switch {
		case event.Type == messages.SEND_MESSAGE:
			{
				url := event.Url

				apiGtw := connection.NewApiGatewayClient(url)

				data, err := json.Marshal(event.Msg)

				if err != nil {
					log.Panicf("error while marshalling message: %s", err.Error())

				}

				dbclient := connection.NewDynamoClient()

				val, err := dynamodbattribute.Marshal(event.To)

				if err != nil {
					log.Panicf("error while marshalling event.To: %s", err.Error())
				}
				scanInput := &dynamodb.ScanInput{
					TableName:        aws.String(os.Getenv("CONNECTION_TABLE")),
					FilterExpression: aws.String("#sortKey = :sortKey"),
					ExpressionAttributeNames: map[string]*string{
						"#sortKey": aws.String("sortKey"),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":sortKey": val,
					},
				}

				output, err := dbclient.Scan(scanInput)

				if err != nil {
					log.Panicf("Scan Error %s", err.Error())
				}

				for _, item := range output.Items {

					defer func() {
						if recover() != nil {
							log.Printf("recovered from: %v \n", item)
						}
					}()

					var dynamoRecord messages.DynamoEvent

					err = dynamodbattribute.UnmarshalMap(item, &dynamoRecord)

					if err != nil {
						log.Panicf("error while unmarshalling records: %s", err.Error())
					}

					_, err = apiGtw.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
						ConnectionId: &dynamoRecord.ConnectionId,
						Data:         data,
					})

					if err != nil {
						log.Panicf("Error while posting, %s", err.Error())
					}

				}

			}
		case event.Type == messages.ENTER_CHAT:
			{
				record := map[string]string{
					"partitionKey": event.ConnectionId,
					"sortKey":      event.From,
				}
				dbclient := connection.NewDynamoClient()

				item, err := dynamodbattribute.MarshalMap(record)

				if err != nil {
					log.Panicf("error while entering chat, %s", err.Error())
				}

				input := &dynamodb.PutItemInput{
					TableName: aws.String(os.Getenv("CONNECTION_TABLE")),
					Item:      item,
				}

				_, err = dbclient.PutItem(input)

				if err != nil {
					log.Panicf("error while entering chat, %s", err.Error())
				}
			}
		case event.Type == messages.BROADCAST:
			{
				url := event.Url

				apiGtw := connection.NewApiGatewayClient(url)

				data, err := json.Marshal(event.Msg)

				if err != nil {
					log.Panicf("error while broadcasting chat, %s", err.Error())
				}

				dbclient := connection.NewDynamoClient()

				scanInput := &dynamodb.ScanInput{
					TableName: aws.String(os.Getenv("CONNECTION_TABLE")),
				}

				output, err := dbclient.Scan(scanInput)

				for _, item := range output.Items {

					go func() {
						if recover() != nil {
							log.Printf("recovered from item error, item: %v", item)
						}
					}()
					var dynamoRecord messages.DynamoEvent

					err = dynamodbattribute.UnmarshalMap(item, &dynamoRecord)

					if err != nil {
						log.Println(err.Error())
					}

					_, err = apiGtw.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
						ConnectionId: &dynamoRecord.ConnectionId,
						Data:         data,
					})

					if err != nil {
						log.Panicf("error while posting message, %s", err.Error())
					}

				}
			}
		case event.Type == messages.LEAVE_CHAT:
			{
				var itemToDelete messages.DynamoEvent
				dynamoConn := connection.NewDynamoClient()
				dynamoKey, err := dynamodbattribute.Marshal(event.ConnectionId)
				if err != nil {
					log.Printf("error while disconnecting %s, event: %v", err.Error(), event)
				}
				// search by id
				{
					input := &dynamodb.QueryInput{
						TableName:              aws.String(os.Getenv("CONNECTION_TABLE")),
						KeyConditionExpression: aws.String("#partitionKey = :partitionKey"),
						ExpressionAttributeNames: map[string]*string{
							"#partitionKey": aws.String("partitionKey"),
						},
						ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
							":partitionKey": dynamoKey,
						},
					}

					output, err := dynamoConn.Query(input)

					if err != nil {
						log.Panicf("error fetching %s", err.Error())
					}

					if len(output.Items) == 0 {
						log.Panicf("no connection found by the id: %s", event.From)
					}

					err = dynamodbattribute.UnmarshalMap(output.Items[0], &itemToDelete)

					if err != nil {
						log.Panicf("error while disconnecting %s \n", err.Error())
					}
					log.Println(itemToDelete)
				}
				// delete
				{
					dynamoKey, err := dynamodbattribute.MarshalMap(messages.DynamoEvent{
						ConnectionId: itemToDelete.ConnectionId,
						To:           itemToDelete.To,
					})

					if err != nil {
						log.Panicf("error while disconnecting %s \n", err.Error())
					}

					log.Println(dynamoKey)

					input := &dynamodb.DeleteItemInput{
						TableName: aws.String(os.Getenv("CONNECTION_TABLE")),
						Key:       dynamoKey,
					}

					_, err = dynamoConn.DeleteItem(input)

					if err != nil {
						log.Panicf("error while deleting item on disconnect %s \n", err.Error())
					}
				}

			}
		default:
			{
				log.Println("unknown event type", event.Type)
			}
		}

		connection := connection.NewSQSClient()

		input := &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(os.Getenv("QUEUE_URL")),
			ReceiptHandle: &record.ReceiptHandle,
		}

		_, err = connection.DeleteMessage(input)

		if err != nil {
			panic("failed to delete message")
		}
	}

	return nil
}
func main() {

	lambda.Start(handler)
}
