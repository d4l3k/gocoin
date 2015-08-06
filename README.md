[![Build Status](https://travis-ci.org/StorjPlatform/gocoin.svg?branch=master)](https://travis-ci.org/StorjPlatform/gocoin)
[![GoDoc](https://godoc.org/github.com/StorjPlatform/gocoin?status.svg)](https://godoc.org/github.com/StorjPlatform/gocoin)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/StorjPlatform/gocoin/master/LICENSE)
[![Coverage Status](https://coveralls.io/repos/StorjPlatform/gocoin/badge.svg?branch=master)](https://coveralls.io/r/StorjPlatform/gocoin?branch=master)


# GOcoin 

##Overview

This is a library to make bitcoin address and transactions which was forked from [hellobitcoin](https://github.com/prettymuchbryce/hellobitcoin).
Additionally you can gether unspent transaction outputs(UTXO) and send transaction by using [Blockr.io](http://blockr.io) WEB API.

This uses btcec library in [btcd](https://github.com/btcsuite/btcd) instead of https://github.com/toxeus/go-secp256k1
not to use C programs.


###Installation

1. Install [go](http://golang.org/)
2. run `go get` to install dependencies


## Example
```go

import gocoin

func main(){
	//make a public and private key pair.
	key, _ := gocoin.GenerateKey(true)
	adr, _ := key.Pub.GetAddress()
	fmt.Println("address=", adr)
	wif := key.Priv.GetWIFAddress()
	fmt.Println("wif=", wif)

	txKey, _ := gocoin.GetKeyFromWIF(wif)


	//gethter unspent trnasactions transaction
	s := gocoin.NewBlockrService(true)
	txs, _ := s.GetUTXO(adr)

	if len(txs) > 1 {
		//create a transaction.
		txin := gocoin.TXin{}
		txin.Hash = txs[0].Hash
		txin.Index = txs[0].Index
		txin.Sequence = uint32(0xffffffff)
		txin.TxPrevScript = txs[0].Script
		txin.key = txKey

		txout := gocoin.TXout{}
		txout.Value = txs[0].Amount - 1000000
		txout.CreateStandardScript("n2eMqTT929pb1RDNuqEnxdaLau1rxy3efi")
		tx := gocoin.TX{}
		tx.Txin = []*gocoin.TXin{&txin}
		tx.Txout = []*gocoin.TXout{&txout}
		tx.Locktime = 0

		rawtx, _:= tx.MakeTX()
		fmt.Println(hex.EncodeToString(rawtx))

	    //send a transaction
		txHash, _:= s.SendTX(rawtx)
		fmt.Println(hex.EncodeToString(txHash))
	}
}
````



# Contribution
Improvements to the codebase and pull requests are encouraged.


