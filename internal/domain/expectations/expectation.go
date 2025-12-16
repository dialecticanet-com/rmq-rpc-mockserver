package expectations

import (
	"time"

	"github.com/google/uuid"
)

type Expectation struct {
	ID         uuid.UUID
	Request    *Request
	Response   *Response
	Times      *Times
	TimeToLive *TimeToLive
	Priority   int
	CreatedAt  time.Time
}

type Times struct {
	RemainingTimes uint32
	Unlimited      bool
}

func (t *Times) Copy() *Times {
	if t == nil {
		return nil
	}

	return &Times{
		RemainingTimes: t.RemainingTimes,
		Unlimited:      t.Unlimited,
	}
}

type TimeToLive struct {
	TTL time.Duration
}

func (t *TimeToLive) Copy() *TimeToLive {
	if t == nil {
		return nil
	}

	return &TimeToLive{
		TTL: t.TTL,
	}
}

func NewExpectation(req *Request, res *Response, opts ...ExpectationOption) (*Expectation, error) {
	e := &Expectation{
		ID:       uuid.New(),
		Request:  req,
		Response: res,
		Priority: 0,
		Times: &Times{
			RemainingTimes: 1,
		},
		CreatedAt: time.Now(),
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	return e, nil
}

func (e *Expectation) Matches(cnd *Candidate) bool {
	return e.IsActive() && e.Request.Matches(cnd)
}

func (e *Expectation) Use() {
	if e.Times != nil && !e.Times.Unlimited {
		e.Times.RemainingTimes--
	}
}

func (e *Expectation) IsActive() bool {
	if e.Times != nil && !e.Times.Unlimited && e.Times.RemainingTimes <= 0 {
		return false
	}

	if e.TimeToLive != nil && time.Now().After(e.CreatedAt.Add(e.TimeToLive.TTL)) {
		return false
	}

	return true
}

func (e *Expectation) Copy() *Expectation {
	return &Expectation{
		ID:         e.ID,
		Request:    e.Request,  // immutable
		Response:   e.Response, // immutable
		Times:      e.Times.Copy(),
		TimeToLive: e.TimeToLive.Copy(),
		Priority:   e.Priority,
		CreatedAt:  e.CreatedAt,
	}
}
