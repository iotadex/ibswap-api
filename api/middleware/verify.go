package middleware

import (
	"encoding/hex"
	"fmt"
	"ibswap/gl"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

func VerifySignature(c *gin.Context) {
	//get user's public key
	sign := c.Query("sign")
	ts := c.Query("ts")
	address := c.Query("address")

	signature, err := hex.DecodeString(strings.TrimPrefix(sign, "0x"))
	if err != nil {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "invalid sign",
		})
		return
	}

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

	if len(address) != 42 {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SIGN_ERROR,
			"err-msg":  "address invalid",
		})
		return
	}

	tsData := []byte(ts)
	ts = fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(tsData), tsData)
	hash := crypto.Keccak256Hash([]byte(ts))
	if err = verifyEthAddress(address, signature, hash.Bytes()); err != nil {
		gl.OutLogger.Error("User's sign error. %s: %s : %s : %v", address, c.Query("ts"), sign, err)
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SIGN_ERROR,
			"err-msg":  "sign error",
		})
		return
	}

	c.Set("account", address)
	c.Next()
}

func verifyEthAddress(address string, signature, hashData []byte) error {
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
	if address != crypto.PubkeyToAddress(*sigPublicKey).Hex() {
		return fmt.Errorf("sign address error")
	}
	return nil
}
