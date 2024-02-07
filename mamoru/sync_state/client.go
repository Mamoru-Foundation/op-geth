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
	"strconv"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

const DeltaBlocks = 10 // min diff between currentBlock and highestBlock

type SyncClient interface {
	CurrentBlock() uint64
	HighestBlock() uint64
	IsSync() bool
}

// JSONRPCRequest represents a JSON-RPC request payload
type JSONRPCRequest struct {
	JsonRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Result  *ResultStruct `json:"result"`
	Error   *RPCError     `json:"error"`
}

type ResultStruct struct {
	CurrentL1          Block `json:"current_l1"`
	CurrentL1Finalized Block `json:"current_l1_finalized"`
	HeadL1             Block `json:"head_l1"`
	SafeL1             Block `json:"safe_l1"`
	FinalizedL1        Block `json:"finalized_l1"`
	UnsafeL2           Block `json:"unsafe_l2"`
	SafeL2             Block `json:"safe_l2"`
	FinalizedL2        Block `json:"finalized_l2"`
	PendingSafeL2      Block `json:"pending_safe_l2"`
}

type Block struct {
	Hash           string   `json:"hash"`
	Number         int      `json:"number"`
	ParentHash     string   `json:"parentHash"`
	Timestamp      int      `json:"timestamp"`
	L1Origin       L1Origin `json:"l1origin,omitempty"`
	SequenceNumber int      `json:"sequenceNumber,omitempty"`
}

type L1Origin struct {
	Hash   string `json:"hash"`
	Number int    `json:"number"`
}

func (r *ResultStruct) GetCurrentBlockNumber() uint64 {
	return uint64(r.SafeL2.Number)
}

func (r *ResultStruct) GetHighestBlockNumber() uint64 {
	return uint64(r.FinalizedL2.Number)
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Client struct {
	Url        string
	PolishTime uint

	syncData *ResultStruct

	quit    chan struct{}
	signals chan os.Signal
}

var _ SyncClient = (*Client)(nil)

func (c *Client) CurrentBlock() uint64 {
	if c.syncData != nil {
		return c.syncData.GetCurrentBlockNumber()
	}
	return 0
}

func (c *Client) HighestBlock() uint64 {
	if c.syncData != nil {
		return c.syncData.GetHighestBlockNumber()
	}
	return 0
}

func (c *Client) IsSync() bool {
	return int64(c.HighestBlock())-int64(c.CurrentBlock()) < getDeltaBlocks()
}

func NewJSONRPCRequest(url string, PolishTime uint) SyncClient {
	c := &Client{
		Url:        url,
		PolishTime: PolishTime,
		quit:       make(chan struct{}),
		signals:    make(chan os.Signal, 1),
	}

	// Register for SIGINT (Ctrl+C) and SIGTERM (kill) signals
	signal.Notify(c.signals, syscall.SIGINT, syscall.SIGTERM)

	go c.loop()

	return c
}

func (c *Client) loop() {
	ticker := time.NewTicker(time.Duration(c.PolishTime) * time.Second)
	defer ticker.Stop()
	// Perform the first tick immediately
	time.Sleep(5 * time.Minute)
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
		return
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

func getDeltaBlocks() int64 {
	delta := os.Getenv("MAMORU_SNIFFER_DELTA")
	if delta == "" {
		return DeltaBlocks
	}

	deltaBlocks, err := strconv.ParseInt(delta, 10, 64)
	if err != nil {
		return DeltaBlocks
	}

	return deltaBlocks
}
