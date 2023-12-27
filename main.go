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

/*
tokens := map[string]string{
		"0xd2a67f5C808C7F525845975423F9979809188E44": "APE",
		"0x174d211F46994860500DB4d66B3EFE605A82BF95": "AUR",
		"0x68acf9Da768a5f43c7D29999D44e9e39026DDDc4": "CS",
		"0x3bBb9B7848De06778fEE4fE0bC4d9AB271e56648": "IHG",
		"0xd459140CFA38f3488F2076c3A4e9271Cf2f2E40C": "OBS",
		"0x1cDF3F46DbF8Cf099D218cF96A769cea82F75316": "sBTC",
		"0xE9308Bf2d95d11E324E0C62FF24bBD4bbc5dA546": "SDDT",
		"0xa158A39d00C79019A01A6E86c56E96C461334Eb0": "sETH",
		"0x1426116752d65111278c9e598E80E3B055D8D571": "SHIMMERINU",
		"0x5dA63f4456A56a0c5Cb0B2104a3610D5CA3d48E8": "sIOTA",
		"0xc5759E47b0590146675C560163036C302Fa05bC3": "SPHE",
		"0x3C844FB5AD27A078d945dDDA8076A4084A76E513": "sSOON",
		"0xc0E49f8C615d3d4c245970F6Dc528E4A47d69a44": "USDT",
		"0x6C890075406C5DF08b427609E3A2eAD1851AD68D": "WSMR",
	}

	type NFT struct {
		User      string `json:"user"`
		Pool      string `json:"pool"`
		Liquidity string `json:"liquidity"`
		L         *big.Int
	}

	rpcClient, err := ethclient.Dial("https://json-rpc.evm.shimmer.network")
	if err != nil {
		panic(fmt.Errorf("dial node error. %v", err))
	}
	nft, err := service.NewINonfungiblePositionManager(common.HexToAddress("0x5f0E8A90f8093aBddF0cA21898B2A71350754a0D"), rpcClient)
	if err != nil {
		panic(err)
	}
	id := int64(0)
	nfts := make(map[string]map[string]*NFT, 0)
	for id < 2055 {
		id++
		time.Sleep(time.Second)
		p, err := nft.Positions(&bind.CallOpts{}, big.NewInt(id))
		if err != nil {
			fmt.Println(id, err)
			continue
		}

		if _, exist := tokens[p.Token0.Hex()]; !exist {
			continue
		}
		if _, exist := tokens[p.Token1.Hex()]; !exist {
			continue
		}
		if p.Liquidity.Sign() == 0 {
			continue
		}
		pair := tokens[p.Token0.Hex()] + "-" + tokens[p.Token1.Hex()]

		user, err := nft.OwnerOf(&bind.CallOpts{}, big.NewInt(id))
		if err != nil {
			fmt.Println(id, err)
			continue
		}

		fmt.Println(id, user.Hex(), pair, p.Liquidity)
		if _, exist := nfts[user.Hex()]; !exist {
			nfts[user.Hex()] = make(map[string]*NFT)
		}
		if _, exist := nfts[user.Hex()][pair]; !exist {
			nfts[user.Hex()][pair] = &NFT{
				User:      user.Hex(),
				Pool:      pair,
				Liquidity: p.Liquidity.String(),
				L:         p.Liquidity,
			}
		} else {
			nfts[user.Hex()][pair].L.Add(nfts[user.Hex()][pair].L, p.Liquidity)
			nfts[user.Hex()][pair].Liquidity = nfts[user.Hex()][pair].L.String()
		}
	}

	// create a file
	file, err := os.Create("result.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// initialize csv writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"user", "pool", "liquidity"})
	for user := range nfts {
		for pool := range nfts[user] {
			n := nfts[user][pool]
			data := []string{n.User, n.Pool, n.Liquidity}
			writer.Write(data)
		}
	}

	return
*/
