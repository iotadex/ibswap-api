package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestEVMSign(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	data := []byte("1655714635")
	hash := crypto.Keccak256Hash(data)
	signature, _ := crypto.Sign(hash.Bytes(), privateKey)
	t.Log(hexutil.Encode(signature))
}

func TestIOTASign(t *testing.T) {
	privateKey, _ := hex.DecodeString("4f4b376e64ac07fab72e76d79bfe8b958541f366887d3a595dcbe971680f0ad2e30c1f106286bd8f2258d326a91ea3b54c8360f1bc99cbfab512538a88bbd17d")
	data := []byte("1655714635")
	sig := ed25519.Sign(privateKey, data)
	t.Log(hex.EncodeToString(sig))
	if !ed25519.Verify(privateKey[32:], data, sig) {
		t.Error("verify error.")
	}
}

func TestVerifyEVMSign(t *testing.T) {
	msg := []byte("1662029116")
	data := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)
	hash := crypto.Keccak256Hash([]byte(data))
	signature, _ := hexutil.Decode("0x6d20ee4b16235036d406ad2e7d4e5718b002fc9b40e08a76548b84446fb29fb776f00e7b4763a3487adbf5223bcbb703d8c4c6809703436c3e80d9761401f7251b")
	signature[64] -= 27
	sigPublicKey, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(crypto.PubkeyToAddress(*sigPublicKey).Hex())
	}

	publiKey, _ := crypto.Ecrecover(hash.Bytes(), signature)
	signatureNoRecoverID := signature[:len(signature)-1] // remove recovery id
	verified := crypto.VerifySignature(publiKey, hash.Bytes(), signatureNoRecoverID)
	t.Log(verified)
}

func TestVerifyIOtaSign(t *testing.T) {
	pubKey, _ := hex.DecodeString("b34ff7a4fc2ea1c3720aed2f07f8da727ff903b30b848b4861c74d56590d8c6f")
	hashData := []byte("1659840567")
	signature, err := hexutil.Decode("0x387d16b88fa13b5896e0dd4af7ad7c9e59b6a78b4f0485810ea08788a54957e25840fa712be81acd92cab44e70796fc441fd0a6a6cbb8342996389670972db04")
	if !ed25519.Verify(pubKey, hashData, signature) {
		t.Log("sign error", err)
	}
}
