package grpc

import (
	"encoding/json"
	"time"

	grpcApi "github.com/dialecticanet-com/rmq-rpc-mockserver/api/v1"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/comparators"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/expectations"
	"github.com/dialecticanet-com/rmq-rpc-mockserver/internal/domain/subscriptions"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

func newSubscription(sub *subscriptions.Subscription) *grpcApi.Subscription {
	return &grpcApi.Subscription{
		Id:    sub.ID().String(),
		Queue: sub.Queue(),
	}
}

func newProtoExpectation(exp *expectations.Expectation) *grpcApi.Expectation {
	expDTO := &grpcApi.Expectation{
		Id:        exp.ID.String(),
		Request:   newProtoRequest(exp.Request),
		Response:  newProtoResponse(exp.Response),
		CreatedAt: exp.CreatedAt.Format(time.RFC3339),
	}

	if exp.Times != nil {
		if exp.Times.Unlimited {
			expDTO.Times = &grpcApi.Times{
				Times: &grpcApi.Times_Unlimited{
					Unlimited: true,
				},
			}
		} else {
			expDTO.Times = &grpcApi.Times{
				Times: &grpcApi.Times_RemainingTimes{
					RemainingTimes: exp.Times.RemainingTimes,
				},
			}
		}
	}

	if exp.TimeToLive != nil {
		expiresAt := exp.CreatedAt.Add(exp.TimeToLive.TTL).Format(time.RFC3339)
		expDTO.ExpiresAt = &expiresAt
	}

	return expDTO
}

func newProtoRequest(req *expectations.Request) *grpcApi.Request {
	protoReq := &grpcApi.Request{
		Exchange:   req.Exchange,
		RoutingKey: req.RoutingKey,
	}

	switch b := req.BodyComparator.(type) {
	case *comparators.JSONBody:
		var v map[string]interface{}
		if err := json.Unmarshal(b.Body, &v); err != nil {
			return nil
		}

		pbValue, err := structpb.NewStruct(v)
		if err != nil {
			return nil
		}

		protoReq.Body = &grpcApi.Request_JsonBody{
			JsonBody: &grpcApi.JSONBodyAssertion{
				Body:      pbValue,
				MatchType: newProtoMatchType(b.MatchType),
			},
		}
	case *comparators.Regex:
		protoReq.Body = &grpcApi.Request_RegexBody{
			RegexBody: &grpcApi.RegexBodyAssertion{
				Regex: b.Regex.String(),
			},
		}
	}

	return protoReq
}

func newProtoMatchType(mt comparators.MatchType) grpcApi.JSONBodyAssertion_MatchType {
	switch mt {
	case comparators.MatchTypeExact:
		return grpcApi.JSONBodyAssertion_MATCH_TYPE_EXACT
	case comparators.MatchTypePartial:
		return grpcApi.JSONBodyAssertion_MATCH_TYPE_PARTIAL
	default:
		return grpcApi.JSONBodyAssertion_MATCH_TYPE_UNSPECIFIED
	}
}

func newProtoResponse(res *expectations.Response) *grpcApi.Response {
	var v interface{}
	if err := json.Unmarshal(res.Body, &v); err != nil {
		return nil
	}

	pbValue, err := structpb.NewValue(v)
	if err != nil {
		return nil
	}

	return &grpcApi.Response{Body: pbValue}
}

func newProtoAssertion(assertion *expectations.Assertion, include []string) *grpcApi.Assertion {
	protoAssertion := &grpcApi.Assertion{
		Id: uuid.New().String(), // Generate new ID for the assertion
		Candidate: &grpcApi.Assertion_Candidate{
			Exchange:   assertion.Candidate.Exchange,
			RoutingKey: assertion.Candidate.RoutingKey,
		},
		CreatedAt: assertion.CreatedAt.Format(time.RFC3339),
	}

	// Convert candidate body to proto value
	var v map[string]interface{}
	if err := json.Unmarshal(assertion.Candidate.Body, &v); err != nil {
		return nil
	}
	pbValue, err := structpb.NewStruct(v)
	if err != nil {
		return nil
	}
	protoAssertion.Candidate.Body = pbValue

	if assertion.Expectation != nil {
		protoAssertion.Matched = true
	}

	// Set matched expectation if exists and if "expectation" is in the include array
	includeExpectation := false
	for _, inc := range include {
		if inc == "expectation" {
			includeExpectation = true
			break
		}
	}

	if includeExpectation && assertion.Expectation != nil {
		protoAssertion.Expectation = newProtoExpectation(assertion.Expectation)
	}

	return protoAssertion
}
