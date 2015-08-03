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

package bitgoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/StorjPlatform/bitgoin/base58check"
)

//	flag.StringVar(&flagPrivateKey, "private-key", "", "The private key of the bitcoin wallet which contains the bitcoins you wish to send.")
//	flag.StringVar(&flagPublicKey, "public-key", "", "The public address of the bitcoin wallet which contains the bitcoins you wish to send.")
//	flag.StringVar(&flagDestination, "destination", "", "The public address of the bitcoin wallet to which you wish to send the bitcoins.")
//	flag.StringVar(&flagInputTransaction, "input-transaction", "", "An unspent input transaction hash which contains the bitcoins you wish to send. (Note: HelloBitcoin assumes a single input transaction, and a single output transaction for simplicity.)")
//	flag.IntVar(&flagInputIndex, "input-index", 0, "The output index of the unspent input transaction which contains the bitcoins you wish to send. Defaults to 0 (first index).")
//	flag.IntVar(&flagSatoshis, "satoshis", 0, "The number of bitcoins you wish to send as represented in satoshis (100,000,000 satoshis = 1 bitcoin). (Important note: the number of satoshis left unspent in your input transaction will be spent as the transaction fee.)")

//MakeTX makes transaction and return tx hex string(not send)
func MakeTX(key *Key, flagDestination, flagInputTransaction string, flagInputIndex, flagSatoshis int) (string, error) {
	//This transaction code is not completely robust.
	//It expects that you have exactly 1 input transaction, and 1 output address.
	//It also expects that your transaction is a standard Pay To Public Key Hash (P2PKH) transaction.
	//This is the most common form used to send a transaction to one or multiple Bitcoin addresses.

	//First we create the raw transaction.
	//In order to construct the raw transaction we need the input transaction hash,
	//the destination address, the number of satoshis to send, and the scriptSig
	//which is temporarily (prior to signing) the ScriptPubKey of the input transaction.
	publicKeyBase58, _ := key.Pub.GetAddress()

	tempScriptSig, err := createScriptPubKey(publicKeyBase58)
	if err != nil {
		return "", err
	}

	rawTransaction, err := createRawTransaction(flagInputTransaction, flagInputIndex, flagDestination, flagSatoshis, tempScriptSig)
	if err != nil {
		return "", err
	}

	//After completing the raw transaction, we append
	//SIGHASH_ALL in little-endian format to the end of the raw transaction.
	hashCodeType, err := hex.DecodeString("01000000")
	if err != nil {
		return "", err
	}

	var rawTransactionBuffer bytes.Buffer
	rawTransactionBuffer.Write(rawTransaction)
	rawTransactionBuffer.Write(hashCodeType)
	rawTransactionWithHashCodeType := rawTransactionBuffer.Bytes()

	//Sign the raw transaction, and output it to the console.
	scriptSig := signRawTransaction(rawTransactionWithHashCodeType, key)
	finalTransaction, err := createRawTransaction(flagInputTransaction, flagInputIndex, flagDestination, flagSatoshis, scriptSig)
	if err != nil {
		return "", err
	}

	finalTransactionHex := hex.EncodeToString(finalTransaction)

	fmt.Println("Your final transaction is")
	fmt.Println(finalTransactionHex)

	return finalTransactionHex, nil
}

func createScriptPubKey(publicKeyBase58 string) ([]byte, error) {
	publicKeyBytes, _, err := base58check.Decode(publicKeyBase58)
	if err != nil {
		return nil, err
	}

	var scriptPubKey bytes.Buffer
	scriptPubKey.WriteByte(byte(118))                 //OP_DUP
	scriptPubKey.WriteByte(byte(169))                 //OP_HASH160
	scriptPubKey.WriteByte(byte(len(publicKeyBytes))) //PUSH
	scriptPubKey.Write(publicKeyBytes)
	scriptPubKey.WriteByte(byte(136)) //OP_EQUALVERIFY
	scriptPubKey.WriteByte(byte(172)) //OP_CHECKSIG
	return scriptPubKey.Bytes(), nil
}

func signRawTransaction(rawTransaction []byte, key *Key) []byte {
	//Here we start the process of signing the raw transaction.

	publicKeyBytes := key.Pub.key.SerializeUncompressed()

	//Hash the raw transaction twice before the signing
	shaHash := sha256.New()
	shaHash.Write(rawTransaction)
	hash := shaHash.Sum(nil)

	shaHash2 := sha256.New()
	shaHash2.Write(hash)
	rawTransactionHashed := shaHash2.Sum(nil)

	//Sign the raw transaction
	sig, err := key.Priv.key.Sign(rawTransactionHashed)
	if err != nil {
		log.Fatal("Failed to sign transaction")
	}
	signedTransaction := sig.Serialize()

	//Verify that it worked.
	verified := sig.Verify(rawTransactionHashed, key.Pub.key)
	if !verified {
		log.Fatal("Failed to sign transaction")
	}

	hashCodeType, err := hex.DecodeString("01")
	if err != nil {
		log.Fatal(err)
	}

	//+1 for hashCodeType
	signedTransactionLength := byte(len(signedTransaction) + 1)

	var publicKeyBuffer bytes.Buffer
	publicKeyBuffer.Write(publicKeyBytes)
	pubKeyLength := byte(len(publicKeyBuffer.Bytes()))

	var buffer bytes.Buffer
	buffer.WriteByte(signedTransactionLength)
	buffer.Write(signedTransaction)
	buffer.WriteByte(hashCodeType[0])
	buffer.WriteByte(pubKeyLength)
	buffer.Write(publicKeyBuffer.Bytes())

	return buffer.Bytes()

	//Return the final transaction
}

func createRawTransaction(inputTransactionHash string, inputTransactionIndex int, publicKeyBase58Destination string, satoshis int, scriptSig []byte) ([]byte, error) {
	//Create the raw transaction.

	//Version field
	version, err := hex.DecodeString("01000000")
	if err != nil {
		return nil, err
	}

	//# of inputs (always 1 in our case)
	inputs, err := hex.DecodeString("01")
	if err != nil {
		return nil, err
	}

	//Input transaction hash
	inputTransactionBytes, err := hex.DecodeString(inputTransactionHash)
	if err != nil {
		return nil, err
	}

	//Convert input transaction hash to little-endian form
	inputTransactionBytesReversed := make([]byte, len(inputTransactionBytes))
	for i := 0; i < len(inputTransactionBytes); i++ {
		inputTransactionBytesReversed[i] = inputTransactionBytes[len(inputTransactionBytes)-i-1]
	}

	//Output index of input transaction
	outputIndexBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(outputIndexBytes, uint32(inputTransactionIndex))

	//Script sig length
	scriptSigLength := len(scriptSig)

	//sequence_no. Normally 0xFFFFFFFF. Always in this case.
	sequence, err := hex.DecodeString("ffffffff")
	if err != nil {
		return nil, err
	}

	//Numbers of outputs for the transaction being created. Always one in this example.
	numOutputs, err := hex.DecodeString("01")
	if err != nil {
		return nil, err
	}

	//Satoshis to send.
	satoshiBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(satoshiBytes, uint64(satoshis))

	//Script pub key
	scriptPubKey, err := createScriptPubKey(publicKeyBase58Destination)
	if err != nil {
		return nil, err
	}

	scriptPubKeyLength := len(scriptPubKey)

	//Lock time field
	lockTimeField, err := hex.DecodeString("00000000")
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.Write(version)
	buffer.Write(inputs)
	buffer.Write(inputTransactionBytesReversed)
	buffer.Write(outputIndexBytes)
	buffer.WriteByte(byte(scriptSigLength))
	buffer.Write(scriptSig)
	buffer.Write(sequence)
	buffer.Write(numOutputs)
	buffer.Write(satoshiBytes)
	buffer.WriteByte(byte(scriptPubKeyLength))
	buffer.Write(scriptPubKey)
	buffer.Write(lockTimeField)

	return buffer.Bytes(), nil
}
