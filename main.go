package main

import (
	"encoding/hex"
	"fmt"
	"ibswap/api"
	"ibswap/daemon"
	"ibswap/gl"
	"ibswap/model"
	"ibswap/service"
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

	api.StartHttpServer()

	service.Start()

	daemon.WaitForKill()

	api.StopHttpServer()
}

func ComputerAddress() {
	/*
			pool = address(
		            uint256(
		                keccak256(
		                    abi.encodePacked(
		                        hex"ff",
		                        factory,
		                        keccak256(abi.encode(key.token0, key.token1, key.fee)),
		                        POOL_INIT_CODE_HASH
		                    )
		                )
		            )
		        );
				var EventBurnV3 = crypto.Keccak256Hash([]byte("Burn(address,int24,int24,uint128,uint256,uint256)"))
	*/
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

//0xfff5505f34d30ecd03811dafc8326874e74900ee767be0ca804036440740d78231d38c626feb4f324dcfd0d0f20b49722870d86de38e9fe382501507411f52f566085a224b4f7955fab4a5dec1aedf007cb94dffb8
//0xfff5505f34d30ecd03811dafc8326874e74900ee767be0ca804036440740d78231d38c626feb4f324dcfd0d0f20b49722870d86de38e9fe382501507411f52f566085a224b4f7955fab4a5dec1aedf007cb94dffb8
//0x99381366B094Cb94e88423A5cF604CFe536793dA
//0x99381366B094Cb94e88423A5cF604CFe536793dA
/*
0000000000000000000000000000000000000000000000000de0b6b3a7640000
ffffffffffffffffffffffffffffffffffffffffffffffffa9ea031bdfffd2b1
00000000000000000000000000000000000000027a3e3a47b53185372cc1e2c7
00000000000000000000000000000000000000000000000685b4e147a248c4f5
00000000000000000000000000000000000000000000000000000000000046e2
*/
