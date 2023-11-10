package sync_state

import "github.com/ethereum/go-ethereum"

type SyncProcess struct {
	client *Client
}

func NewSyncProcess(opNodeRpcUrl string, PolishTime uint) *SyncProcess {
	return &SyncProcess{
		client: NewJSONRPCRequest(opNodeRpcUrl, PolishTime),
	}
}

func (s SyncProcess) Start() {
	go s.client.loop()
}

func (s SyncProcess) Stop() {
	s.client.Close()
}

func (s SyncProcess) Progress() ethereum.SyncProgress {
	return ethereum.SyncProgress{
		CurrentBlock: s.client.syncData.GetCurrentBlockNumber(),
		HighestBlock: s.client.syncData.GetHighestBlockNumber(),
	}
}
