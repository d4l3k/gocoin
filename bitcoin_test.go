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
 */

package gocoin

import (
	"fmt"
	"testing"
)

func TestBitcoin(t *testing.T) {
	key, err := GenerateKey(true)
	if err != nil {
		t.Errorf(err.Error())
	}
	adr, _ := key.Pub.GetAddress()
	fmt.Println("address=", adr)
	wif := key.Priv.GetWIFAddress()
	fmt.Println("wif=", wif)

	key2, err := GetKeyFromWIF(wif)
	if err != nil {
		t.Errorf(err.Error())
	}
	adr2, _ := key2.Pub.GetAddress()
	fmt.Println("address2=", adr2)

	if adr != adr2 {
		t.Errorf("key unmatched")
	}

	wif = "928Qr9J5oAC6AYieWJ3fG3dZDjuC7BFVUqgu4GsvRVpoXiTaJJf"
	txKey, err := GetKeyFromWIF(wif)
	if err != nil {
		t.Errorf(err.Error())
	}
	adr, _ = txKey.Pub.GetAddress()
	fmt.Println("address for tx=", adr)
	if adr!="n3Bp1hbgtmwDtjQTpa6BnPPCA8fTymsiZy"{
		t.Errorf("invalid address")
	}
		
	tx, err := MakeTX(txKey, "n2eMqTT929pb1RDNuqEnxdaLau1rxy3efi", "1a103718e2e0462c50cb057a0f39d7c6cbf960276452d07dc4a50ddca725949c", 1, 68000000)
	ok := "01000000019c9425a7dc0da5c47dd052642760f9cbc6d7390f7a05cb502c46e0e21837101a010000008a473044022030ebb89d54e76b9e14b8eb21aa30055eb54289dcd3aad9b415ebcc153b211eee0220720fa77cfc2c25da52899f3bf9a947869bc89d26066c02a1c428e9530a3f49b10141049f160b18fa4acedccdc063961d63b3a23385b1e67159d07521cb46d4e7209ecd443e473796e7ace130164c660fbcfb7dcac8437cc55f3ceafb546054c8d8cbdfffffffff0100990d04000000001976a914e7c1345fc8f87c68170b3aa798a956c2fe6a9eff88ac00000000"
	if tx != ok {
		t.Errorf("invalid tx")
	}
	fmt.Println(tx)


	key, err = GenerateKey(false)
	if err != nil {
		t.Errorf(err.Error())
	}
	adr, _ = key.Pub.GetAddress()
	fmt.Println("address=", adr)
	wif = key.Priv.GetWIFAddress()
	fmt.Println("wif=", wif)

	key2, err = GetKeyFromWIF(wif)
	if err != nil {
		t.Errorf(err.Error())
	}
	adr2, _ = key2.Pub.GetAddress()
	fmt.Println("address2=", adr2)

	if adr != adr2 {
		t.Errorf("key unmatched")
	}

}
