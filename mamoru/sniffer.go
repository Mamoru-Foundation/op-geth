package mamoru

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/Mamoru-Foundation/mamoru-sniffer-go/mamoru_sniffer"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/mamoru/sync_state"
)

var (
	connectMutex       sync.Mutex
	sniffer            *mamoru_sniffer.Sniffer
	SnifferConnectFunc = mamoru_sniffer.Connect
)

const DefaultPolishTime = 10

var syncProcess sync_state.SyncClient

func mapToInterfaceSlice(m map[string]string) []interface{} {
	var result []interface{}
	for key, value := range m {
		result = append(result, key, value)
	}

	return result
}

func init() {
	mamoru_sniffer.InitLogger(func(entry mamoru_sniffer.LogEntry) {
		kvs := mapToInterfaceSlice(entry.Ctx)
		msg := "Mamoru core: " + entry.Message
		switch entry.Level {
		case mamoru_sniffer.LogLevelDebug:
			log.Debug(msg, kvs...)
		case mamoru_sniffer.LogLevelInfo:
			log.Info(msg, kvs...)
		case mamoru_sniffer.LogLevelWarning:
			log.Warn(msg, kvs...)
		case mamoru_sniffer.LogLevelError:
			log.Error(msg, kvs...)
		}
	})
	opNodeUrl, ok := os.LookupEnv("MAMORU_OP_NODE_URL")
	if ok {
		polishTimeEnv := os.Getenv("MAMORU_OP_NODE_POLISH_TIME_SEC")
		var polishTime uint = DefaultPolishTime
		if polishTimeEnv != "" {
			parseUint, err := strconv.ParseUint(polishTimeEnv, 10, 32)
			if err != nil {
				log.Error("Mamoru OP NODE POLISH TIME parse error", "err", err)
				parseUint = DefaultPolishTime
			}
			polishTime = uint(parseUint)
		}
		InitSyncProcess(func() (sync_state.SyncClient, error) {
			return sync_state.NewJSONRPCRequest(opNodeUrl, polishTime), nil
		})
	}
}

type Sniffer struct {
}

func NewSniffer() *Sniffer {
	return &Sniffer{}
}

func (s *Sniffer) CheckRequirements(currentBlock uint64) bool {
	return s.isSnifferEnable() && s.checkSynced(int64(currentBlock)) && connect()
}

func (s *Sniffer) checkSynced(currentBlock int64) bool {
	if syncProcess == nil {
		return false
	}

	if currentBlock == 0 {
		return false
	}

	isSync := syncProcess.IsSync()

	highestBlock := int64(syncProcess.HighestBlock())
	log.Info("Mamoru Sniffer sync", "syncing", isSync, "CurrentBlock", currentBlock, "highest", highestBlock, "diff", highestBlock-currentBlock)

	return isSync
}

func (s *Sniffer) isSnifferEnable() bool {
	val, ok := os.LookupEnv("MAMORU_SNIFFER_ENABLE")
	if !ok {
		return false
	}

	isEnable, err := strconv.ParseBool(val)
	if err != nil {
		log.Error("Mamoru Sniffer env parse error", "err", err)
		return false
	}

	return isEnable
}

func InitConnectFunc(f func() (*mamoru_sniffer.Sniffer, error)) {
	SnifferConnectFunc = f
}

func connect() bool {
	connectMutex.Lock()
	defer connectMutex.Unlock()

	if sniffer != nil {
		return true
	}

	var err error
	if sniffer == nil {
		sniffer, err = SnifferConnectFunc()
		if err != nil {
			erst := strings.Replace(err.Error(), "\t", "", -1)
			erst = strings.Replace(erst, "\n", "", -1)
			//	erst = strings.Replace(erst, " ", "", -1)
			log.Error("Mamoru Sniffer connect", "err", erst)

			return false
		}
	}

	return true
}

func InitSyncProcess(f func() (sync_state.SyncClient, error)) {
	var err error
	syncProcess, err = f()
	if err != nil {
		log.Error("Mamoru Sniffer sync process init error", "err", err)
	}
}
