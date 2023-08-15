package connection

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func newSession() *session.Session {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return sess
}
func NewApiGatewayClient(domain string) *apigatewaymanagementapi.ApiGatewayManagementApi {

	sess := newSession()

	client := apigatewaymanagementapi.New(sess, &aws.Config{
		Endpoint: aws.String(domain),
	})

	return client
}

func NewSQSClient() *sqs.SQS {

	sess := newSession()
	svc := sqs.New(sess)

	return svc
}

func NewDynamoClient() *dynamodb.DynamoDB {
	return dynamodb.New(newSession())
}
