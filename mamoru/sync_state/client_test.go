package sync_state

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockSyncStatusResponse = `{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "current_l1": {
      "hash": "0x2ed093ae8a0639f39a96f28d6485920df569c87f3a2afc0ae30a0300ae4f2bc3",
      "number": 4280269,
      "parentHash": "0xd707a7c18998c62c80c29f6c1ebd6740c8834ab5f7d10135796650aaa0a8d221",
      "timestamp": 1694607492
    },
    "current_l1_finalized": {
      "hash": "0x901fc921673992cbeb925565cfa7fd955b767421358be11293b26f7f8313d464",
      "number": 5808982,
      "parentHash": "0x697dc2a419b47a809c7a11ce5d3f504797f0c55b7580509a6e833119a6847c1d",
      "timestamp": 1714484064
    },
    "head_l1": {
      "hash": "0x365369b37043f93e57e24186bb05915d76256870dd09e8fd062e9352ca42fdfc",
      "number": 5809065,
      "parentHash": "0x4846a4d8a99054beab9a397ef4f6d726649bfd52bceeb5d80714f09fcdf57bf1",
      "timestamp": 1714485240
    },
    "safe_l1": {
      "hash": "0x0a0d059eea56c2492beffa2603b3cc0da2d52aae7d3343403a0b1a9152e4b0b5",
      "number": 5809009,
      "parentHash": "0xdd626defb2bbdf2038931617610a453390ef2aea6996c0e6a80083c018454db6",
      "timestamp": 1714484448
    },
    "finalized_l1": {
      "hash": "0x901fc921673992cbeb925565cfa7fd955b767421358be11293b26f7f8313d464",
      "number": 5808982,
      "parentHash": "0x697dc2a419b47a809c7a11ce5d3f504797f0c55b7580509a6e833119a6847c1d",
      "timestamp": 1714484064
    },
    "unsafe_l2": {
      "hash": "0xddad2d0456bbd5b6a0106462f5a0d7fdfa6cfff9f6a5a5eafaacd12dc42e1ced",
      "number": 1402442,
      "parentHash": "0xea513790d2446a2803b234220ee6317c13682fa2121880136cde3511399928da",
      "timestamp": 1694607424,
      "l1origin": {
        "hash": "0xb731e912e14d4f09b22d85432d12c9e14fc3ee33c40d24a3bcc084010b44336e",
        "number": 4280258
      },
      "sequenceNumber": 1
    },
    "safe_l2": {
      "hash": "0xddad2d0456bbd5b6a0106462f5a0d7fdfa6cfff9f6a5a5eafaacd12dc42e1ced",
      "number": 1402442,
      "parentHash": "0xea513790d2446a2803b234220ee6317c13682fa2121880136cde3511399928da",
      "timestamp": 1694607424,
      "l1origin": {
        "hash": "0xb731e912e14d4f09b22d85432d12c9e14fc3ee33c40d24a3bcc084010b44336e",
        "number": 4280258
      },
      "sequenceNumber": 1
    },
    "finalized_l2": {
      "hash": "0xddad2d0456bbd5b6a0106462f5a0d7fdfa6cfff9f6a5a5eafaacd12dc42e1ced",
      "number": 1402442,
      "parentHash": "0xea513790d2446a2803b234220ee6317c13682fa2121880136cde3511399928da",
      "timestamp": 1694607424,
      "l1origin": {
        "hash": "0xb731e912e14d4f09b22d85432d12c9e14fc3ee33c40d24a3bcc084010b44336e",
        "number": 4280258
      },
      "sequenceNumber": 1
    },
    "pending_safe_l2": {
      "hash": "0xddad2d0456bbd5b6a0106462f5a0d7fdfa6cfff9f6a5a5eafaacd12dc42e1ced",
      "number": 1402442,
      "parentHash": "0xea513790d2446a2803b234220ee6317c13682fa2121880136cde3511399928da",
      "timestamp": 1694607424,
      "l1origin": {
        "hash": "0xb731e912e14d4f09b22d85432d12c9e14fc3ee33c40d24a3bcc084010b44336e",
        "number": 4280258
      },
      "sequenceNumber": 1
    }
  }
}`

// mockServer simulates an Ethereum node server for testing
func mockServer(responseBody string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}

// TestSendJSONRPCRequestSuccess tests the sendJSONRPCRequest function for a successful scenario
func TestSendJSONRPCRequestSuccess(t *testing.T) {
	server := mockServer(mockSyncStatusResponse, http.StatusOK)
	defer server.Close()

	requestPayload := NewSyncStatusRequest()

	response, err := sendJSONRPCRequest(context.Background(), server.URL, requestPayload)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Errorf("Expected response non-nil response")
		return
	}

	var result = response.Result
	if result == nil {
		t.Errorf("Expected response.Result  non-nil response")
		return
	}

	assert.Equal(t, uint64(1402442), result.GetCurrentBlockNumber())
	assert.Equal(t, uint64(1402442), result.GetHighestBlockNumber())
}

// TestSendJSONRPCRequestError tests the sendJSONRPCRequest function for a scenario with an error response
func TestSendJSONRPCRequestError(t *testing.T) {
	server := mockServer(`{"jsonrpc":"2.0","id":1,"error":{"code":-32600,"message":"Invalid Request"}}`, http.StatusOK)
	defer server.Close()

	requestPayload := NewSyncStatusRequest()

	response, err := sendJSONRPCRequest(context.Background(), server.URL, requestPayload)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response.Error == nil {
		t.Errorf("Expected an error response")
	}
}

func Test_getDeltaBlocks(t *testing.T) {
	t.Run("Success return delta from env ", func(t *testing.T) {
		want := int64(100)
		_ = os.Setenv("MAMORU_SNIFFER_DELTA", fmt.Sprintf("%d", want))
		defer os.Unsetenv("MAMORU_SNIFFER_DELTA")
		got := getDeltaBlocks()
		assert.Equal(t, want, got)
	})
	t.Run("Success return delta from env ", func(t *testing.T) {
		want := int64(DeltaBlocks)
		got := getDeltaBlocks()
		assert.Equal(t, want, got)
		defer os.Unsetenv("MAMORU_SNIFFER_DELTA")
	})
}
