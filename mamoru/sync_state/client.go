package sync_state

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

// JSONRPCRequest represents a JSON-RPC request payload
type JSONRPCRequest struct {
	JsonRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// BlockStatus represents the status of a block
type BlockStatus struct {
	Hash       string `json:"hash"`
	Number     uint64 `json:"number"`
	ParentHash string `json:"parentHash"`
	Timestamp  int    `json:"timestamp"`
}

// L1Origin represents the origin of a block on L1
type L1Origin struct {
	Hash   string `json:"hash"`
	Number int    `json:"number"`
}

// L2BlockStatus extends BlockStatus with L1 origin and sequence number
type L2BlockStatus struct {
	BlockStatus
	L1Origin       L1Origin `json:"l1origin"`
	SequenceNumber int      `json:"sequenceNumber"`
}

// SyncStatusResponse represents the response for the sync status
type SyncStatusResponse struct {
	CurrentL1          BlockStatus   `json:"current_l1"`
	CurrentL1Finalized BlockStatus   `json:"current_l1_finalized"`
	FinalizedL2        L2BlockStatus `json:"finalized_l2"`
	EngineSyncTarget   L2BlockStatus `json:"engine_sync_target"`
	QueuedUnsafeL2     L2BlockStatus `json:"queued_unsafe_l2"`
	// Add other fields as needed...
}

func (sync *SyncStatusResponse) GetCurrentBlockNumber() uint64 {
	return sync.EngineSyncTarget.Number
}

func (sync *SyncStatusResponse) GetHighestBlockNumber() uint64 {
	return sync.QueuedUnsafeL2.Number
}

// JSONRPCResponse represents a JSON-RPC response payload
type JSONRPCResponse struct {
	JsonRPC string             `json:"jsonrpc"`
	ID      int                `json:"id"`
	Result  SyncStatusResponse `json:"result"`
	Error   *RPCError          `json:"error"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Client struct {
	syncData   SyncStatusResponse
	Url        string
	PolishTime uint

	quit    chan struct{}
	signals chan os.Signal
}

func NewJSONRPCRequest(url string, PolishTime uint) *Client {
	c := &Client{
		Url:        url,
		PolishTime: PolishTime,
		quit:       make(chan struct{}),
		signals:    make(chan os.Signal, 1),
	}

	// Register for SIGINT (Ctrl+C) and SIGTERM (kill) signals
	signal.Notify(c.signals, syscall.SIGINT, syscall.SIGTERM)

	return c
}

func (c *Client) loop() {
	ticker := time.NewTicker(time.Duration(c.PolishTime) * time.Second)
	defer ticker.Stop()
	// Perform the first tick immediately
	c.fetchSyncStatus()

	for {
		select {
		case <-ticker.C:
			c.fetchSyncStatus()
		case <-c.quit:
			log.Info("Mamoru SyncProcess Shutting down...")
			return
		case <-c.signals:
			log.Info("Signal received, initiating shutdown...")
			c.Close()
		}
	}
}
func NewSyncStatusRequest() JSONRPCRequest {
	return JSONRPCRequest{
		JsonRPC: "2.0",
		Method:  "optimism_syncStatus",
		Params:  []interface{}{},
		ID:      1,
	}
}
func (c *Client) fetchSyncStatus() {
	// Set up the request payload
	requestPayload := NewSyncStatusRequest()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Info("Mamoru requesting syncData ...")
	// Send the request and get the response
	response, err := sendJSONRPCRequest(ctx, c.Url, requestPayload)
	if err != nil {
		log.Error("Mamoru Sync", "err", err)
		return
	}

	// Process the response
	if response.Error != nil {
		log.Error("Mamoru Sync", "err", response.Error.Message, "url", c.Url, "code", response.Error.Code)
	}

	c.syncData = response.Result
}

func (c *Client) Close() {
	close(c.quit)
}

func sendJSONRPCRequest(ctx context.Context, url string, requestPayload JSONRPCRequest) (*JSONRPCResponse, error) {
	// Marshal the request payload to JSON
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create an HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the HTTP POST request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var response JSONRPCResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response body: %w", err)
	}

	return &response, nil
}
