package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var cfg aws.Config

func NewApiGatewayClient(domain string) *apigatewaymanagementapi.Client {

	client := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = &domain
	})

	return client
}

func NewDynamoClient() *dynamodb.Client {
	return dynamodb.NewFromConfig(cfg)
}

func init() {
	c, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	cfg = c
}
