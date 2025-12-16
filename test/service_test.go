package test

import (
	"context"
	"net/http"
	"testing"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	gocoreamqp "github.com/dialecticanet-com/rmq-rpc-mockserver/lib/amqp"
	gocoregrpc "github.com/dialecticanet-com/rmq-rpc-mockserver/lib/grpc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestGRPCFlow(t *testing.T) {
	ctx := context.Background()
	NewHTTPExpect(t).DELETE("/api/v1/reset").Expect().Status(http.StatusOK)

	// setup grpc connection to a server
	cl, err := gocoregrpc.NewInsecureClient(serviceHostGRPC)
	require.NoError(t, err)

	gRPCClient := grpcApi.NewAmqpMockServerServiceClient(cl)

	// create a queue
	queue, routingKey := createRandomQueue(t)
	subResp, err := gRPCClient.AddSubscription(ctx, &grpcApi.AddSubscriptionRequest{
		Queue: queue,
	})
	require.NoError(t, err)
	require.NoError(t, uuid.Validate(subResp.GetSubscription().GetId()))

	req := &grpcApi.CreateExpectationRequest{
		Request: &grpcApi.Request{
			Exchange:   testExchange,
			RoutingKey: routingKey,
			Body: &grpcApi.Request_JsonBody{
				JsonBody: &grpcApi.JSONBodyAssertion{
					MatchType: grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL,
					Body:      newJSONBodyAsStruct(t, `{"fooCamel": "bar", "baz_snake": "qux"}`),
				},
			},
		},
		Response: &grpcApi.Response{Body: newJSONBodyAsValue(t, `{"oneCamel":"val1","sec_snake":"val2"}`)},
	}

	// call the grpc server
	resp, err := gRPCClient.CreateExpectation(ctx, req)
	require.NoError(t, err)
	require.NoError(t, uuid.Validate(resp.GetExpectationId()))

	rpcClient, err := gocoreamqp.NewRPCClient(rmqCon)
	require.NoError(t, err)

	// call the rpc server
	qres, err := rpcClient.Call(ctx, testExchange, routingKey, []byte(`{"fooCamel": "bar", "baz_snake": "qux", "extra": "data"}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"oneCamel":"val1", "sec_snake":"val2"}`, string(qres))
}

func TestHTTPFlow(t *testing.T) {
	ctx := context.Background()
	NewHTTPExpect(t).DELETE("/api/v1/reset").Expect().Status(http.StatusOK)

	queue, routingKey := createRandomQueue(t)

	NewHTTPExpect(t).POST("/api/v1/subscriptions").WithJSON(grpcApi.AddSubscriptionRequest{
		Queue: queue,
	}).Expect().Status(http.StatusOK).JSON().Object()

	req := &grpcApi.CreateExpectationRequest{
		Request: &grpcApi.Request{
			Exchange:   testExchange,
			RoutingKey: routingKey,
			Body: &grpcApi.Request_JsonBody{
				JsonBody: &grpcApi.JSONBodyAssertion{
					MatchType: grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL,
					Body:      newJSONBodyAsStruct(t, `{"fooCamel1": "bar", "baz_snake": "qux"}`),
				},
			},
		},
		Response: &grpcApi.Response{Body: newJSONBodyAsValue(t, `{"oneCamel":"val11","sec_snake":"val21"}`)},
	}
	reqBody, err := protojson.Marshal(req)
	require.NoError(t, err)

	NewHTTPExpect(t).POST("/api/v1/expectations").WithBytes(reqBody).Expect().Status(http.StatusOK).JSON().Object()

	rpcClient, err := gocoreamqp.NewRPCClient(rmqCon)
	require.NoError(t, err)

	// call the rpc server
	qres, err := rpcClient.Call(ctx, testExchange, routingKey, []byte(`{"fooCamel1": "bar", "baz_snake": "qux", "extra": "data"}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"oneCamel":"val11", "sec_snake":"val21"}`, string(qres))
}

func TestQueuesSubscriptions(t *testing.T) {
	NewHTTPExpect(t).DELETE("/api/v1/reset").Expect().Status(http.StatusOK)

	queue1, _ := createRandomQueue(t)
	queue2, _ := createRandomQueue(t)
	queue3, _ := createRandomQueue(t)

	subID1 := NewHTTPExpect(t).POST("/api/v1/subscriptions").WithJSON(grpcApi.AddSubscriptionRequest{
		Queue: queue1,
	}).Expect().Status(http.StatusOK).JSON().Object().Value("id").String().Raw()
	require.NotEmpty(t, subID1)

	subID2 := NewHTTPExpect(t).POST("/api/v1/subscriptions").WithJSON(grpcApi.AddSubscriptionRequest{
		Queue: queue2,
	}).Expect().Status(http.StatusOK).JSON().Object().Value("id").String().Raw()
	require.NotEmpty(t, subID2)

	// idempotent subscription
	subID21 := NewHTTPExpect(t).POST("/api/v1/subscriptions").WithJSON(grpcApi.AddSubscriptionRequest{
		Queue:      queue2,
		Idempotent: true,
	}).Expect().Status(http.StatusOK).JSON().Object().Value("id").String().Raw()
	require.Equal(t, subID2, subID21)

	subID3 := NewHTTPExpect(t).POST("/api/v1/subscriptions").WithJSON(grpcApi.AddSubscriptionRequest{
		Queue: queue3,
	}).Expect().Status(http.StatusOK).JSON().Object().Value("id").String().Raw()
	require.NotEmpty(t, subID3)

	NewHTTPExpect(t).GET("/api/v1/subscriptions").Expect().Status(http.StatusOK).JSON().
		Array().Length().IsEqual(3)

	NewHTTPExpect(t).DELETE("/api/v1/subscriptions/" + subID1).Expect().Status(http.StatusOK)
	NewHTTPExpect(t).DELETE("/api/v1/subscriptions/queues/" + queue2).Expect().Status(http.StatusOK)

	NewHTTPExpect(t).GET("/api/v1/subscriptions").Expect().Status(http.StatusOK).JSON().
		Array().Length().IsEqual(1)

	NewHTTPExpect(t).GET("/api/v1/subscriptions").Expect().Status(http.StatusOK).JSON().
		Array().Value(0).Object().Value("id").IsEqual(subID3)
}
