package treesql

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/phayes/freeport"
)

func NewTestServer() (*Server, *ClientConn, error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(dir)

	port := freeport.GetPort()

	server := NewServer(dir+"/test.data", port)
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	url := fmt.Sprintf("ws://localhost:%d/ws", port)
	client, err := NewClientConn(url)
	if err != nil {
		return nil, nil, err
	}

	return server, client, nil
}

// define stmt => define error or ack
// define query => define error or initialResponse
type simpleTestStmt struct {
	stmt  string
	query string

	ack           string
	error         string
	initialResult string
}

// runSimpleTestScript spins up a test server and runs statements on it,
// checking each result. It doesn't support live queries; only initial results
// are checked.
func runSimpleTestScript(t *testing.T, cases []simpleTestStmt) {
	server, client, err := NewTestServer()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	defer server.Close()

	for idx, testCase := range cases {
		// Run a statement.
		if testCase.stmt != "" {
			result, err := client.Exec(testCase.stmt)
			assertError(t, idx, testCase.error, err)
			if result != testCase.ack {
				t.Fatalf(`case %d: expected ack "%s"; got "%s"`, idx, testCase.ack, result)
			}
			continue
		}
		// Run a query.
		if testCase.query != "" {
			res, err := client.Query(testCase.query)
			assertError(t, idx, testCase.error, err)
			indented, _ := json.MarshalIndent(res.Data, "", "  ")
			if string(indented) != testCase.initialResult {
				t.Fatalf("expected:\n%sgot:\n%s", testCase.initialResult, indented)
			}
		}
	}
}

func assertError(t *testing.T, caseIdx int, expected string, err error) {
	if err != nil {
		if expected == "" {
			t.Fatalf(`case %d: expected success; got error "%s"`, caseIdx, err.Error())
		}
		if err.Error() != expected {
			t.Fatalf(`case %d: expected error "%s"; got "%s"`, caseIdx, expected, err.Error())
		}
	}
	if err == nil && expected != "" {
		t.Fatalf(`case %d: expected error "%s"; got success`, caseIdx, expected)
	}
}