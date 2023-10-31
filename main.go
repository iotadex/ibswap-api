package main

import (
	"encoding/hex"
	"fmt"
	"ibdex/api"
	"ibdex/api/middleware"
	"ibdex/config"
	"ibdex/daemon"
	"ibdex/gl"
	"ibdex/model"
	"ibdex/service"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "-d")
	}

	daemon.Background("./out.log", true)

	gl.CreateLogFiles()

	model.ConnectToMysql()

	middleware.AddAdmins(config.AdminAddresses)
	api.StartHttpServer()

	service.Start()

	daemon.WaitForKill()

	api.StopHttpServer()
}

func ComputerAddress() {
	t0 := common.HexToAddress("0xc9f3a2C8a5C05FDbE086549de9DD9954ACA7BD22")
	t1 := common.HexToAddress("0xdcC4E969F081C3E967581Aa9175EF6F0a337Ae88")
	fee := big.NewInt(10000)
	factory := common.HexToAddress("0xf5505F34D30eCd03811DAFC8326874e74900Ee76")
	InitCode, _ := hex.DecodeString("8e9fe382501507411f52f566085a224b4f7955fab4a5dec1aedf007cb94dffb8")

	data := common.LeftPadBytes(t0.Bytes(), 32)
	data = append(data, common.LeftPadBytes(t1.Bytes(), 32)...)
	data = append(data, common.LeftPadBytes(fee.Bytes(), 32)...)
	s1 := crypto.Keccak256(data)
	fmt.Println(hex.EncodeToString(s1))

	d := []byte{15 + 15<<4}
	d = append(d, factory.Bytes()...)
	d = append(d, s1...)
	d = append(d, InitCode...)
	s2 := crypto.Keccak256(d)
	fmt.Println(hex.EncodeToString(d))

	pool := common.BytesToAddress(s2[12:])
	fmt.Println(pool.Hex())
}
