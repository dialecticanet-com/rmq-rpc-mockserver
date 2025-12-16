package test

import (
	"context"
	"net/http"
	"testing"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	gocoreamqp "github.com/dialecticanet-com/rmq-rpc-mockserver/lib/amqp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestExpectationsHTTPAPI(t *testing.T) {
	ctx := context.Background()

	// Reset all expectations before the test
	NewHTTPExpect(t).DELETE("/api/v1/reset").Expect().Status(http.StatusOK)

	// Create a random queue and routing key
	queue, routingKey := createRandomQueue(t)

	// Create a subscription for the queue
	subResp := NewHTTPExpect(t).POST("/api/v1/subscriptions").WithJSON(grpcApi.AddSubscriptionRequest{
		Queue: queue,
	}).Expect().Status(http.StatusOK).JSON().Object()

	// Verify subscription ID is valid
	subID := subResp.Value("id").String().Raw()
	require.NoError(t, uuid.Validate(subID))

	// Create an expectation
	req := &grpcApi.CreateExpectationRequest{
		Request: &grpcApi.Request{
			Exchange:   testExchange,
			RoutingKey: routingKey,
			Body: &grpcApi.Request_JsonBody{
				JsonBody: &grpcApi.JSONBodyAssertion{
					MatchType: grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL,
					Body:      newJSONBodyAsStruct(t, `{"foo": "bar", "baz": "qux"}`),
				},
			},
		},
		Response: &grpcApi.Response{Body: newJSONBodyAsValue(t, `{"result":"success","data":"test-data"}`)},
	}
	reqBody, err := protojson.Marshal(req)
	require.NoError(t, err)

	// Send the expectation creation request
	expResp := NewHTTPExpect(t).POST("/api/v1/expectations").WithBytes(reqBody).
		Expect().Status(http.StatusOK).JSON().Object()

	// Verify expectation ID is valid
	expID := expResp.Value("expectation_id").String().Raw()
	require.NoError(t, uuid.Validate(expID))

	// Get all expectations
	expList := NewHTTPExpect(t).GET("/api/v1/expectations").
		Expect().Status(http.StatusOK).JSON().Object().
		Value("expectations").Array()

	// Verify there is exactly one expectation
	expList.Length().IsEqual(1)

	// should have zero expired expectation
	NewHTTPExpect(t).GET("/api/v1/expectations").WithQuery("status", "expired").
		Expect().Status(http.StatusOK).JSON().Object().Value("expectations").Array().Length().IsEqual(0)

	// should have one active expectation
	NewHTTPExpect(t).GET("/api/v1/expectations").WithQuery("status", "active").
		Expect().Status(http.StatusOK).JSON().Object().Value("expectations").Array().Length().IsEqual(1)

	// Verify the expectation has the correct ID
	expList.Value(0).Object().Value("id").IsEqual(expID)

	// Get the specific expectation by ID
	getExp := NewHTTPExpect(t).GET("/api/v1/expectations/" + expID).
		Expect().Status(http.StatusOK).JSON().Object()

	// Verify the expectation details
	getExp.Value("id").IsEqual(expID)
	getExp.Value("request").Object().Value("exchange").IsEqual(testExchange)
	getExp.Value("request").Object().Value("routing_key").IsEqual(routingKey)

	// Test the expectation by sending a message to RabbitMQ
	rpcClient, err := gocoreamqp.NewRPCClient(rmqCon)
	require.NoError(t, err)

	// Call the RPC server with a message that matches the expectation
	qres, err := rpcClient.Call(ctx, testExchange, routingKey, []byte(`{"foo": "bar", "baz": "qux", "extra": "field"}`))
	require.NoError(t, err)

	// Verify the response matches the expected response
	assert.JSONEq(t, `{"result":"success","data":"test-data"}`, string(qres))

	// should have one expired expectation
	NewHTTPExpect(t).GET("/api/v1/expectations").WithQuery("status", "expired").
		Expect().Status(http.StatusOK).JSON().Object().Value("expectations").Array().Length().IsEqual(1)

	// should have zero active expectations
	NewHTTPExpect(t).GET("/api/v1/expectations").WithQuery("status", "active").
		Expect().Status(http.StatusOK).JSON().Object().Value("expectations").Array().Length().IsEqual(0)
}
