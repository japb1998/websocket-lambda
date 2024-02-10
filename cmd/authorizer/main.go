package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var h = slog.NewTextHandler(os.Stdout, nil).WithAttrs([]slog.Attr{slog.String("pkg", "authorizer")})
var logger = slog.New(h)

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	logger.Info("Authorizer event", slog.String("event", fmt.Sprintf("%+v", event)))
	token, ok := event.QueryStringParameters["Auth"]

	if !ok || token == "" {
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	} else {
		resource := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/$connect", os.Getenv("REGION"), os.Getenv("ACCOUNT_ID"), os.Getenv("API_ID"), os.Getenv("STAGE"))
		logger.Info("resource", slog.String("event", event.Resource), slog.String("resource", resource))

		// METHOD arn should be arn:aws:execute-api:region:account-id:api-id/stage-name/$connect
		return generatePolicy("user", "Allow", fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/$connect", os.Getenv("REGION"), os.Getenv("ACCOUNT_ID"), os.Getenv("API_ID"), os.Getenv("STAGE")), token), nil
	}
}
func main() {
	lambda.Start(handler)
}

// generatePolicy creates a policy for the given principalId, effect and resource
func generatePolicy(principalId, effect, resource, token string) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}
	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}
	// any property you want to pass to the next handler
	authResponse.Context = map[string]interface{}{
		"email": token,
	}
	return authResponse
}
