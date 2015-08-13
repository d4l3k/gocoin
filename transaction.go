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
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/StorjPlatform/gocoin/base58check"
)

//TXin represents tx input of a transaction.
type TXin struct {
	Hash             []byte
	Index            uint32
	scriptSig        []byte
	Sequence         uint32
	PrevScriptPubkey []byte
	CreateScriptSig  func(rawTransactionHashed []byte) ([]byte, error)
}

//TXout represents tx output of a transaction.
type TXout struct {
	Value  uint64
	Script []byte
}

//TX represents transaction.
type TX struct {
	Txin     []*TXin
	Txout    []*TXout
	Locktime uint32
}

//MakeTX makes transaction and return tx hex string(not send)
func (tx *TX) MakeTX() ([]byte, error) {
	var err error
	for i, in := range tx.Txin {
		rawTransactionHashed := tx.getRawTransactionHash(i)

		if in.CreateScriptSig == nil {
			return nil, errors.New("in.CreateScriptSig must be set")
		}
		in.scriptSig, err = in.CreateScriptSig(rawTransactionHashed[:])
		if err != nil {
			return nil, err
		}
	}
	//Sign the raw transaction, and output it to the console.
	finalTransaction := tx.createRawTransaction(-1)
	finalTransactionHex := hex.EncodeToString(finalTransaction)

	logging.Println("Your final transaction is")
	logging.Println(finalTransactionHex)

	return finalTransaction, nil
}

//CreateStandardScriptPubkey creates standard script pubkey .
func CreateStandardScriptPubkey(publicKeyBase58 string) ([]byte, error) {
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
func CreateStandardScriptSig(rawTransactionHashed []byte, key *Key) ([]byte, error) {

	publicKeyBytes := key.Pub.key.SerializeUncompressed()

	//Sign the raw transaction
	sig, err := key.Priv.key.Sign(rawTransactionHashed)
	if err != nil {
		return nil, errors.New("failed to sign transaction")
	}
	signedTransaction := sig.Serialize()

	//Verify that it worked.
	verified := sig.Verify(rawTransactionHashed[:], key.Pub.key)
	if !verified {
		return nil, errors.New("Failed to sign transaction")
	}

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

	scriptSig := buffer.Bytes()

	return scriptSig, nil
	//Return the final transaction
}

func (tx *TX) getRawTransactionHash(numSign int) []byte {
	rawTransaction := tx.createRawTransaction(numSign)
	//After completing the raw transaction, we append
	//SIGHASH_ALL in little-endian format to the end of the raw transaction.
	hashCodeType := []byte{0x01, 0x00, 0x00, 0x00}

	var rawTransactionBuffer bytes.Buffer
	rawTransactionBuffer.Write(rawTransaction)
	rawTransactionBuffer.Write(hashCodeType)
	rawTransactionWithHashCodeType := rawTransactionBuffer.Bytes()
	//Hash the raw transaction twice before the signing
	hash := sha256.Sum256(rawTransactionWithHashCodeType)
	h := sha256.Sum256(hash[:])
	return h[:]
}

//createRawTransaction creates a transaction from tx struct.
//if numSing>=0, this returns a transaction for singning, and
//numSign is number of txin which will be singed later.
//if numSing<0, returns a transaction for broadcast.
func (tx *TX) createRawTransaction(numSign int) []byte {
	//Create the raw transaction.
	var buffer bytes.Buffer

	//Version field
	version := []byte{0x01, 0x00, 0x00, 0x00}
	buffer.Write(version)

	//# of inputs (always 1 in our case)
	inputs := toVI(uint64(len(tx.Txin)))
	buffer.Write(inputs)

	for nIn, in := range tx.Txin {
		//Input transaction hash

		//Convert input transaction hash to little-endian form
		inputTransactionBytesReversed := make([]byte, len(in.Hash))
		for i, tb := range in.Hash {
			inputTransactionBytesReversed[len(in.Hash)-i-1] = tb
		}
		buffer.Write(inputTransactionBytesReversed)

		//Output index of input transaction
		outputIndexBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(outputIndexBytes, in.Index)
		buffer.Write(outputIndexBytes)

		var script []byte
		switch {
		case nIn == numSign:
			script = in.PrevScriptPubkey
		case numSign >= 0:
			script = []byte{}
		default:
			script = in.scriptSig
		}
		//Script sig length
		scriptSigLength := len(script)
		buffer.Write(toVI(uint64(scriptSigLength)))

		buffer.Write(script)

		//sequence_no. Normally 0xFFFFFFFF. Always in this case.
		seqBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(seqBytes, in.Sequence)
		buffer.Write(seqBytes)
	}

	//# of outputs (always 1 in our case)
	outputs := toVI(uint64(len(tx.Txout)))
	buffer.Write(outputs)

	for _, out := range tx.Txout {
		//Satoshis to send.
		satoshiBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(satoshiBytes, out.Value)
		buffer.Write(satoshiBytes)

		//Script sig length
		scriptSigLength := len(out.Script)
		buffer.Write(toVI(uint64(scriptSigLength)))

		buffer.Write(out.Script)
	}

	//Lock time field
	lockTimeField := make([]byte, 4)
	binary.LittleEndian.PutUint32(lockTimeField, tx.Locktime)
	buffer.Write(lockTimeField)

	return buffer.Bytes()
}

func toVI(n uint64) []byte {
	if n < uint64(0xfd) {
		b := make([]byte, 1)
		b[0] = byte(n & 0xff)
		return b
	}
	if n <= uint64(0xffff) {
		b := make([]byte, 3)
		b[0] = 0xfd
		binary.LittleEndian.PutUint16(b[1:], uint16(n))
		return b
	}
	if n <= uint64(0xffffffff) {
		b := make([]byte, 5)
		b[0] = 0xfe
		binary.LittleEndian.PutUint32(b[1:], uint32(n))
		return b
	}
	b := make([]byte, 9)
	b[0] = 0xff
	binary.LittleEndian.PutUint64(b[1:], n)
	return b
}
