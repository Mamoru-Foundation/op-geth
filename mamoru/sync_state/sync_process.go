package sync_state

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/log"
)

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
	current := s.client.syncData.GetCurrentBlockNumber()
	highest := s.client.syncData.GetHighestBlockNumber()
	log.Info("Mamoru SyncProcess", "current", current, "highest", highest)
	return ethereum.SyncProgress{
		CurrentBlock: current,
		HighestBlock: highest,
	}
}
