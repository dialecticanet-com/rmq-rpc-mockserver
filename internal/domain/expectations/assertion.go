package expectations

import "time"

type Assertion struct {
	Candidate   *Candidate
	Expectation *Expectation // can be null if no match
	CreatedAt   time.Time
}

func NewMatchedAssertion(cnd *Candidate, exp *Expectation) *Assertion {
	return &Assertion{
		Candidate:   cnd,
		Expectation: exp.Copy(), // freeze the state of the expectation
		CreatedAt:   time.Now(),
	}
}

func NewUnmatchedAssertion(cnd *Candidate) *Assertion {
	return &Assertion{
		Candidate: cnd,
		CreatedAt: time.Now(),
	}
}

type Assertions struct {
	list []*Assertion
}

func NewAssertions() *Assertions {
	return &Assertions{}
}

func (a *Assertions) Add(assertion *Assertion) {
	a.list = append(a.list, assertion)
}

func (a *Assertions) Clear() {
	a.list = nil
}

func (a *Assertions) GetUnmatched() []*Assertion {
	var unmatched []*Assertion
	for _, assertion := range a.list {
		if assertion.Expectation == nil {
			unmatched = append(unmatched, assertion)
		}
	}

	return unmatched
}

func (a *Assertions) GetMatched() []*Assertion {
	var matched []*Assertion
	for _, assertion := range a.list {
		if assertion.Expectation != nil {
			matched = append(matched, assertion)
		}
	}

	return matched
}

func (a *Assertions) GetAll() []*Assertion {
	return a.list
}

func (a *Assertions) GetForExpectationByID(id string) []*Assertion {
	var assertions []*Assertion
	for _, assertion := range a.list {
		if assertion.Expectation != nil && assertion.Expectation.ID.String() == id {
			assertions = append(assertions, assertion)
		}
	}

	return assertions
}
