[![Build Status](https://travis-ci.org/StorjPlatform/gocoin.svg?branch=master)](https://travis-ci.org/StorjPlatform/gocoin)
[![GoDoc](https://godoc.org/github.com/StorjPlatform/gocoin?status.svg)](https://godoc.org/github.com/StorjPlatform/gocoin)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/StorjPlatform/gocoin/master/LICENSE)
[![Coverage Status](https://coveralls.io/repos/StorjPlatform/gocoin/badge.svg?branch=master)](https://coveralls.io/r/StorjPlatform/gocoin?branch=master)


# GOcoin 

## Overview

This is a library to make bitcoin address and transactions which was forked from [hellobitcoin](https://github.com/prettymuchbryce/hellobitcoin),
and has some additional functions.

This uses btcec library in [btcd](https://github.com/btcsuite/btcd) instead of https://github.com/toxeus/go-secp256k1
to be a pure GO program.


## Functions 

1. Normaly Payment(P2PKH) including multi TXin and multi TXout.
2. Gethering unspent transaction outputs(UTXO) and send transactions by using [Blockr.io](http://blockr.io) WEB API.
3. M of N multisig whose codes was partially ported from https://github.com/soroushjp/go-bitcoin-multisig.


## Installation

    $ mkdir tmp
    $ cd tmp
    $ mkdir src
    $ mkdir bin
    $ mkdir pkg
    $ exoprt GOPATH=`pwd`
    $ go get github.com/StorjPlatform/gocoin


## Example
(This omits error handling for simplicity.)

```go

import gocoin

func main(){
	//make a public and private key pair.
	key, _ := gocoin.GenerateKey(true)
	adr, _ := key.Pub.GetAddress()
	fmt.Println("address=", adr)
	wif := key.Priv.GetWIFAddress()
	fmt.Println("wif=", wif)
	
	//get key from wif
	wif := "928Qr9J5oAC6AYieWJ3fG3dZDjuC7BFVUqgu4GsvRVpoXiTaJJf"
	txKey, _ := gocoin.GetKeyFromWIF(wif)

	txKey, _ := gocoin.GetKeyFromWIF(wif)

	//get unspent transactions
	service := gocoin.NewBlockrService(true)
	txs, _ := service.GetUTXO(adr,nil)
	
	//Normal Payment
	_, err = Pay([]*Key{txKey}, map[string]uint64{"n2eMqTT929pb1RDNuqEnxdaLau1rxy3efi": 0.01 * satoshi}, service)
	
	//2 of 3 multisig
	key1, _ := gocoin.GenerateKey(true)
	key2, _ := gocoin.GenerateKey(true)
	key3, _ := gocoin.GenerateKey(true)
	rs, _:= gocoin.NewRedeemScript(2, []*PublicKey{key1.Pub, key2.Pub, key3.Pub})
	//make a fund
	_, err = rs.Pay([]*Key{txKey}, 5000000, service)

    //get a raw transaction for signing.
	rawtx, tx, _:= rs.CreateRawTransactionHashed(map[string]uint64{"n3Bp1hbgtmwDtjQTpa6BnPPCA8fTymsiZy": 50000*satoshi}, service)

	//spend the fund
	sign1, _:= key2.Priv.Sign(rawtx)
	sign2, _:= key3.Priv.Sign(rawtx)
	_, err = rs.Spend(tx, [][]byte{sign1, sign2}, service)
}
````



# Contribution
Improvements to the codebase and pull requests are encouraged.


