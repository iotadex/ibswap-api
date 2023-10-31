package middleware

import (
	"bytes"
	"fmt"
	"ibdex/gl"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

var AdminAddresses = make(map[string]bool)

func VerifySignature(c *gin.Context) {
	//get user's public key
	sign := c.Query("sign")
	ts := c.Query("ts")
	address := common.HexToAddress(c.Query("address"))

	signature := common.FromHex(sign)
	timeStamp, _ := strconv.ParseInt(ts, 10, 64)
	if (timeStamp + 600) < time.Now().Unix() {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SIGN_ERROR,
			"err-msg":  "sign expired",
		})
		return
	}

	tsData := []byte(ts)
	ts = fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(tsData), tsData)
	hash := crypto.Keccak256Hash([]byte(ts))
	if err := verifyEthAddress(address, signature, hash.Bytes()); err != nil {
		gl.OutLogger.Error("User's sign error. %s: %s : %s : %v", address.Hex(), c.Query("ts"), sign, err)
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SIGN_ERROR,
			"err-msg":  "sign error",
		})
		return
	}
	if !AdminAddresses[address.Hex()] {
		gl.OutLogger.Error("user forbidden. %s", address.Hex())
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SIGN_ERROR,
			"err-msg":  "user forbidden",
		})
		return
	}
	c.Next()
}

func verifyEthAddress(address common.Address, signature, hashData []byte) error {
	if len(signature) < 65 {
		return fmt.Errorf("signature length is too short")
	}
	if signature[64] < 27 {
		if signature[64] != 0 && signature[64] != 1 {
			return fmt.Errorf("signature error")
		}
	} else {
		signature[64] -= 27
	}
	sigPublicKey, err := crypto.SigToPub(hashData, signature)
	if err != nil {
		return fmt.Errorf("sign error")
	}
	if !bytes.Equal(address[:], crypto.PubkeyToAddress(*sigPublicKey).Bytes()) {
		return fmt.Errorf("sign address error")
	}
	return nil
}

func AddAdmins(addresses []string) {
	for _, addr := range addresses {
		AdminAddresses[addr] = true
	}
}
