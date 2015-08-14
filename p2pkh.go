/*
 * Copyright (c) 2015, Shinya Yagyu
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 * 3. Neither the name of the copyright holder nor the names of its
 *    contributors may be used to endorse or promote products derived from this
 *    software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 *
 * See LICENSE file for the original license:
 */

package gocoin

import (
	"bytes"
	"encoding/hex"
	"errors"
	"sort"

	"github.com/StorjPlatform/gocoin/base58check"
)

//CreateStandardScriptPubkey creates standard script pubkey .
func createP2PKHScriptPubkey(publicKeyBase58 string) ([]byte, error) {
	publicKeyBytes, _, err := base58check.Decode(publicKeyBase58)
	if err != nil {
		return nil, err
	}

	var scriptPubKey bytes.Buffer
	scriptPubKey.WriteByte(opDUP)
	scriptPubKey.WriteByte(opHASH160)
	scriptPubKey.WriteByte(byte(len(publicKeyBytes))) //PUSH
	scriptPubKey.Write(publicKeyBytes)
	scriptPubKey.WriteByte(opEQUALVERIFY)
	scriptPubKey.WriteByte(opCHECKSIG)
	script := scriptPubKey.Bytes()
	return script, nil
}

//CreateStandardScript creates standard scriptsig and fills TXin.Script.
func createP2PKHScriptSig(rawTransactionHashed []byte, key *Key) ([]byte, error) {

	publicKeyBytes := key.Pub.key.SerializeUncompressed()

	//Sign the raw transaction
	sig, err := key.Priv.key.Sign(rawTransactionHashed)
	if err != nil {
		return nil, errors.New("failed to sign transaction")
	}
	signedTransaction := sig.Serialize()

	//	//Verify that it worked.
	//	verified := sig.Verify(rawTransactionHashed[:], key.Pub.key)
	//	if !verified {
	//		return nil, errors.New("Failed to sign transaction")
	//	}

	//+1 for hashCodeType
	signedTransactionLength := byte(len(signedTransaction) + 1)

	var publicKeyBuffer bytes.Buffer
	publicKeyBuffer.Write(publicKeyBytes)
	pubKeyLength := byte(len(publicKeyBuffer.Bytes()))

	var buffer bytes.Buffer
	buffer.WriteByte(signedTransactionLength)
	buffer.Write(signedTransaction)
	buffer.WriteByte(0x01) //hashCodeType
	buffer.WriteByte(pubKeyLength)
	buffer.Write(publicKeyBuffer.Bytes())

	return buffer.Bytes(), nil
}

func setupP2PKHTXin(keys []*Key, totalAmount uint64, service Service) ([]*TXin, uint64, error) {
	var utxos UTXOs
	for _, k := range keys {
		adr, _ := k.Pub.GetAddress()
		txs, err := service.GetUTXO(adr, k)
		if err != nil {
			return nil, 0, err
		}
		utxos = append(utxos, txs...)
	}

	sort.Sort(utxos)
	txins := make([]*TXin, 0, 10)
	var amount uint64
	for i := range utxos {
		utxo := utxos[len(utxos)-1-i]
		logging.Println("using utxo", hex.EncodeToString(utxo.Hash))
		txin := TXin{}
		txin.Hash = utxo.Hash
		txin.Index = utxo.Index
		txin.Sequence = uint32(0xffffffff)
		txin.PrevScriptPubkey = utxo.Script
		txin.CreateScriptSig = func(rawTransaction []byte) ([]byte, error) {
			return createP2PKHScriptSig(rawTransaction, utxo.Key)
		}
		txins = append(txins, &txin)
		SetTXSpent(utxo.Hash)
		if amount += utxo.Amount; amount >= totalAmount {
			return txins, amount - totalAmount, nil
		}
	}
	return nil, amount - totalAmount, errors.New("not enough coin")
}

func setupP2PKHTXout(addresses map[string]uint64) ([]*TXout, error) {
	txouts := make([]*TXout, len(addresses))
	var err error
	var i int
	for to, amount := range addresses {
		txouts[i] = &TXout{}
		txouts[i].Value = amount
		txouts[i].ScriptPubkey, err = createP2PKHScriptPubkey(to)
		if err != nil {
			return nil, err
		}
		i++
	}
	return txouts, nil
}

//Pay pays in a nomal way.(P2KSH)
func Pay(keys []*Key, addresses map[string]uint64, service Service) ([]byte, error) {
	var err error

	var totalAmount uint64
	for _, amount := range addresses {
		totalAmount += amount
	}

	tx := TX{}
	tx.Locktime = 0
	var remain uint64
	tx.Txin, remain, err = setupP2PKHTXin(keys, totalAmount+fee, service)
	if err != nil {
		return nil, err
	}

	adr, _ := keys[0].Pub.GetAddress()
	if am, exist := addresses[adr]; exist {
		addresses[adr] = am + remain
	} else {
		addresses[adr] = remain
	}
	tx.Txout, err = setupP2PKHTXout(addresses)

	rawtx, err := tx.MakeTX()
	if err != nil {
		return nil, err
	}
	txHash, err := service.SendTX(rawtx)
	if err != nil {
		return nil, err
	}
	logging.Println("tx hash", hex.EncodeToString(txHash))
	return txHash, nil
}
