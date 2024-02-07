package mempool

import (
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/event"
)

type TxPool interface {
	SubscribeTransactions(ch chan<- core.NewTxsEvent, reorgs bool) event.Subscription
}
