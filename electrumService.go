package gocoin

import (
	"encoding/hex"

	"github.com/d4l3k/go-electrum/electrum"
)

type ElectrumService struct {
	node *electrum.Node
}

//NewElectrumService creates ElectrumService struct for not test.
func NewElectrumService() (Service, error) {
	s := &ElectrumService{}
	s.node = electrum.NewNode()
	if err := s.node.ConnectTCP("electrum.dragonzone.net:50001"); err != nil {
		return nil, err
	}
	return s, nil
}

//GetServiceName return service name.
func (s *ElectrumService) GetServiceName() string {
	return "ElectrumService"
}

//SendTX send a transaction using Electrum.
func (s *ElectrumService) SendTX(data []byte) ([]byte, error) {

	resp, err := s.node.BlockchainTransactionBroadcast(data)
	if err != nil {
		return nil, err
	}
	logging.Printf("%+v", resp)
	return hex.DecodeString(resp)
}

//GetUTXO gets unspent transaction outputs by using Electrum.
func (s *ElectrumService) GetUTXO(addr string, key *Key) (UTXOs, error) {
	if cacheUTXO[addr] != nil {
		return cacheUTXO[addr], nil
	}

	headers, err := s.node.BlockchainHeadersSubscribe()
	if err != nil {
		return nil, err
	}
	txs, err := s.node.BlockchainAddressListUnspent(addr)
	if err != nil {
		return nil, err
	}
	header := <-headers

	utxos := make(UTXOs, 0, len(txs))
	for _, tx := range txs {
		utxo := UTXO{}
		utxo.Addr = addr
		utxo.Age = header.BlockHeight - tx.Height
		utxo.Amount = tx.Value
		utxo.Hash, err = hex.DecodeString(tx.Hash)
		if err != nil {
			return nil, err
		}
		utxo.Key = key
		// Note: electrum doesn't return number of signatures required, nor script info
		utxos = append(utxos, &utxo)
	}
	if key != nil {
		cacheUTXO[addr] = utxos
	}
	return utxos, nil
}
