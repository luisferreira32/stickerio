package test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	api "github.com/luisferreira32/stickerio/api"
)

// TODO: more fixtures and options

func fixtureStartMovement() *http.Request {
	m := api.V1Movement{}
	b, err := m.MarshalJSON()
	if err != nil {
		panic(err)
	}

	r, err := http.NewRequest(
		"POST",
		"http://localhost:8080/v1/movements",
		bytes.NewReader(b),
	)
	if err != nil {
		panic(err)
	}
	return r
}

func Test_regressions(t *testing.T) {
	testcases := []struct {
		name          string
		requestChain  []*http.Request
		responseChain []*http.Response
	}{}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			if len(testcase.requestChain) != len(testcase.responseChain) {
				t.Fatalf("unexpected test format where len(testcase.requestChain) != len(testcase.responseChain)")
			}

			for i := 0; i < len(testcase.requestChain); i++ {
				got, err := http.DefaultClient.Do(testcase.requestChain[i])
				if err != nil {
					t.Errorf("unexpected error on request %d: %v", i, err)
				}

				if diff := cmp.Diff(got, testcase.responseChain[i]); diff != "" {
					t.Errorf("unexpected diff in response: %v", diff)
				}
			}
		})
	}
}
