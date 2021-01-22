package store

import (
	"bytes"
	"github.com/jacoblai/yiblockchain/utils"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

type LvStore struct {
	db           *leveldb.DB
	currentBatch *leveldb.Transaction
}

var _ abcitypes.Application = (*LvStore)(nil)

func NewLvStore(dir string) (*LvStore, error) {
	bloom := utils.Precision(float64(256)*1.44, 0, true)
	opts := &opt.Options{}
	opts.ErrorIfMissing = false
	opts.BlockCacheCapacity = 4 * utils.MB
	opts.Filter = filter.NewBloomFilter(int(bloom))
	opts.Compression = opt.SnappyCompression
	opts.BlockSize = 4 * utils.KB
	opts.WriteBuffer = 4 * utils.MB
	opts.OpenFilesCacheCapacity = 1 * utils.KB
	opts.CompactionTableSize = 32 * utils.MB
	opts.WriteL0SlowdownTrigger = 16
	opts.WriteL0PauseTrigger = 64

	// Open database for the queue.
	db, err := leveldb.OpenFile(dir+"/lv.db", opts)
	if err != nil {
		return nil, err
	}

	return &LvStore{
		db: db,
	}, nil
}

func (LvStore) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{}
}

func (LvStore) SetOption(req abcitypes.RequestSetOption) abcitypes.ResponseSetOption {
	return abcitypes.ResponseSetOption{}
}

func (l *LvStore) Query(reqQuery abcitypes.RequestQuery) abcitypes.ResponseQuery {
	resQuery := abcitypes.ResponseQuery{
		Key: reqQuery.Data,
	}
	val, err := l.db.Get(reqQuery.Data, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			resQuery.Log = "does not exist"
		} else {
			resQuery.Log = err.Error()
		}
	}
	resQuery.Log = "exists"
	resQuery.Value = val
	return resQuery
}

func (l *LvStore) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	parts := bytes.Split(req.Tx, []byte("="))
	if len(parts) != 2 {
		return abcitypes.ResponseCheckTx{Code: 1, GasWanted: 1}
	}
	key, _ := parts[0], parts[1]
	if ok, _ := l.db.Has(key, nil); ok {
		return abcitypes.ResponseCheckTx{Code: 2, GasWanted: 1}
	}
	return abcitypes.ResponseCheckTx{Code: abcitypes.CodeTypeOK, GasWanted: 1}
}

func (LvStore) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}

func (l *LvStore) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	l.currentBatch, _ = l.db.OpenTransaction()
	return abcitypes.ResponseBeginBlock{}
}

func (l *LvStore) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	parts := bytes.Split(req.Tx, []byte("="))
	if len(parts) != 2 {
		return abcitypes.ResponseDeliverTx{Code: 1}
	}
	key, value := parts[0], parts[1]
	if ok, _ := l.db.Has(key, nil); ok {
		return abcitypes.ResponseDeliverTx{Code: 2}
	}
	_ = l.currentBatch.Put(key, value, nil)
	return abcitypes.ResponseDeliverTx{Code: abcitypes.CodeTypeOK}
}

func (l *LvStore) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{}
}

func (l *LvStore) Commit() abcitypes.ResponseCommit {
	_ = l.currentBatch.Commit()
	return abcitypes.ResponseCommit{Data: []byte{}}
}

func (l *LvStore) ListSnapshots(req abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

func (l *LvStore) OfferSnapshot(req abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

func (l *LvStore) LoadSnapshotChunk(req abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (l *LvStore) ApplySnapshotChunk(req abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}
