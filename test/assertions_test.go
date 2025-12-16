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

func TestAssertions(t *testing.T) {
	ctx := context.Background()

	// Reset all expectations before the test
	NewHTTPExpect(t).DELETE("/api/v1/reset").Expect().Status(http.StatusOK)

	// Create a random queue and routing key
	queue, routingKey := createRandomQueue(t)

	// Create a subscription for the queue
	NewHTTPExpect(t).POST("/api/v1/subscriptions").WithJSON(grpcApi.AddSubscriptionRequest{Queue: queue}).
		Expect().Status(http.StatusOK)

	// Create an expectation that will be met
	req := &grpcApi.CreateExpectationRequest{
		Request: &grpcApi.Request{
			Exchange:   testExchange,
			RoutingKey: routingKey,
			Body: &grpcApi.Request_JsonBody{
				JsonBody: &grpcApi.JSONBodyAssertion{
					MatchType: grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL,
					Body:      newJSONBodyAsStruct(t, `{"foo": "baz"}`),
				},
			},
		},
		Response: &grpcApi.Response{Body: newJSONBodyAsValue(t, `{"result":"success","data":"test-data"}`)},
	}
	reqBody, err := protojson.Marshal(req)
	require.NoError(t, err)
	expResp := NewHTTPExpect(t).POST("/api/v1/expectations").WithBytes(reqBody).
		Expect().Status(http.StatusOK).JSON().Object()
	expectationID := expResp.Value("expectation_id").String().Raw()
	require.NoError(t, uuid.Validate(expectationID))

	// create an expectation that will not be met
	req = &grpcApi.CreateExpectationRequest{
		Request: &grpcApi.Request{
			Exchange:   testExchange,
			RoutingKey: routingKey,
			Body: &grpcApi.Request_JsonBody{
				JsonBody: &grpcApi.JSONBodyAssertion{
					MatchType: grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL,
					Body:      newJSONBodyAsStruct(t, `{"foo2": "baz2"}`),
				},
			},
		},
		Response: &grpcApi.Response{Body: newJSONBodyAsValue(t, `{"result":"success","data":"test-data"}`)},
	}
	reqBody, err = protojson.Marshal(req)
	require.NoError(t, err)
	NewHTTPExpect(t).POST("/api/v1/expectations").WithBytes(reqBody).Expect().Status(http.StatusOK)

	// get all assertions, should be empty
	assertionsResp := NewHTTPExpect(t).GET("/api/v1/assertions").Expect().Status(http.StatusOK).JSON().Object()
	assertionsResp.Value("assertions").Array().IsEmpty()

	// call the rpc server
	rpcClient, err := gocoreamqp.NewRPCClient(rmqCon)
	require.NoError(t, err)

	// one matched expectation
	qres, err := rpcClient.Call(ctx, testExchange, routingKey, []byte(`{"foo": "baz"}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"result":"success","data":"test-data"}`, string(qres))

	// one unmatched expectation
	qres, err = rpcClient.Call(ctx, testExchange, routingKey, []byte(`{"foo": "baz_none"}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"errors":"no match found"}`, string(qres))

	t.Run("gets all assertions", func(t *testing.T) {
		// get all assertions
		assertionsResp = NewHTTPExpect(t).GET("/api/v1/assertions").WithQuery("include", "expectation").Expect().Status(http.StatusOK).JSON().Object()
		assertions := assertionsResp.Value("assertions").Array()
		assertions.Length().IsEqual(2)

		// first, matched
		assertion := assertions.Value(0).Object()
		assertion.Value("id").String().NotEmpty()
		assertion.Value("candidate").Object().Value("exchange").String().IsEqual(testExchange)
		assertion.Value("candidate").Object().Value("routing_key").String().IsEqual(routingKey)
		assertion.Value("expectation").Object().Value("id").String().NotEmpty()
		assertion.Value("matched").Boolean().IsTrue()

		// second, unmatched
		assertion = assertions.Value(1).Object()
		assertion.Value("id").String().NotEmpty()
		assertion.Value("candidate").Object().Value("exchange").String().IsEqual(testExchange)
		assertion.Value("candidate").Object().Value("routing_key").String().IsEqual(routingKey)
		assertion.NotContainsKey("expectation")
		assertion.Value("matched").Boolean().IsFalse()
	})

	t.Run("gets all matched assertions", func(t *testing.T) {
		// get only matched assertions
		assertionsResp = NewHTTPExpect(t).GET("/api/v1/assertions").WithQuery("status", "matched").
			WithQuery("include", "expectation").Expect().Status(http.StatusOK).JSON().Object()
		assertions := assertionsResp.Value("assertions").Array()
		assertions.Length().IsEqual(1)

		// first, matched
		assertion := assertions.Value(0).Object()
		assertion.Value("id").String().NotEmpty()
		assertion.Value("expectation").Object().Value("id").String().NotEmpty()
		assertion.Value("matched").Boolean().IsTrue()
	})

	t.Run("gets all unmatched assertions", func(t *testing.T) {
		// get only unmatched assertions
		assertionsResp = NewHTTPExpect(t).GET("/api/v1/assertions").WithQuery("status", "unmatched").
			WithQuery("include", "expectation").Expect().Status(http.StatusOK).JSON().Object()
		assertions := assertionsResp.Value("assertions").Array()
		assertions.Length().IsEqual(1)

		// first, unmatched
		assertion := assertions.Value(0).Object()
		assertion.Value("id").String().NotEmpty()
		assertion.NotContainsKey("expectation")
		assertion.Value("matched").Boolean().IsFalse()
	})

	t.Run("get by expectation id", func(t *testing.T) {
		// get only matched assertions
		assertionsResp = NewHTTPExpect(t).GET("/api/v1/assertions").WithQuery("expectation_id", expectationID).
			WithQuery("include", "expectation").Expect().Status(http.StatusOK).JSON().Object()
		assertions := assertionsResp.Value("assertions").Array()
		assertions.Length().IsEqual(1)

		// first, matched
		assertion := assertions.Value(0).Object()
		assertion.Value("id").String().NotEmpty()
		assertion.Value("expectation").Object().Value("id").IsEqual(expectationID)
	})
}
