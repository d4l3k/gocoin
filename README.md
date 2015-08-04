[![Build Status](https://travis-ci.org/StorjPlatform/gocoin.svg?branch=master)](https://travis-ci.org/StorjPlatform/gocoin)
[![GoDoc](https://godoc.org/github.com/StorjPlatform/gocoin?status.svg)](https://godoc.org/github.com/StorjPlatform/gocoin)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/StorjPlatform/gocoin/master/LICENSE)
[![Coverage Status](https://coveralls.io/repos/StorjPlatform/gocoin/badge.svg?branch=master)](https://coveralls.io/r/StorjPlatform/gocoin?branch=master)


# GOcoin 

##Overview

This is a library to make bitcoin address and transactions which was forked from [hellowcoin](https://github.com/prettymuchbryce/hellobitcoin).

This uses btcec library in [btcd](https://github.com/btcsuite/btcd) instead of https://github.com/toxeus/go-secp256k1
not to use C programs.


###Installation

1. Install [go](http://golang.org/)
2. run `go get` to install dependencies


## Example
```go

import bitgoin

func main(){
	key, _ := bitgoin.GenerateKey(true)
	adr, _ := key.Pub.GetAddress()
	fmt.Println("address=", adr)
	wif := key.Priv.GetWIFAddress()
	fmt.Println("wif=", wif)

	txKey, _ := bitgoin.GetKeyFromWIF(wif)
	txKey, _ := bitgoin.GetKeyFromWIF(wif)
	tx, _ := bitgoin.MakeTX(txKey, "n2eMqTT929pb1RDNuqEnxdaLau1rxy3efi", "1a103718e2e0462c50cb057a0f39d7c6cbf960276452d07dc4a50ddca725949c", 1, 68000000)
	fmt.Println(tx)
}
````



# Contribution
Improvements to the codebase and pull requests are encouraged.


