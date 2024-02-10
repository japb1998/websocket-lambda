package connection

type ConnectionItem struct {
	ConnectionId string `json:"sortKey" dynamodbav:"sortKey"`
	User         string `json:"partitionKey" dynamodbav:"partitionKey"`
}
