# api address
```
    https://dex.iotabee.com
```

# APIs for ibswap contract of V2 and V3

## POST /coins/add
### param
```
symbol: SMR,
chain_id: 1071,
contract: 0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e,
code: smr,
type": 0,
decimal: 18,
public: 1
```
### Respose
```
OK
```

## GET /coins/all
```
Get all the tokens' data.
```
### Respose
```
See below response of /coins/{chain_id}/all
```

## GET coins/{chain_id}/all
```
Get all tokens by chain id
```
### param
* chain_id: the chain id of network
### respose
```json
[
  {
    "symbol": "SMR",
    "chain_id": 1071,
    "contract": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "code": "smr",
    "type": 0,
    "decimal": 18,
    "public": 1
  }
]
```
* type : 0 is evm platform token, 1 is erc20 token
* public: 0 (private) or 1 (public)

## GET /coins/{chain_id}/{contract}
```
Get a token by chain id and contract
```
### param
* chain_id: the chain id of network
* contract: the address of token
### respose
```json
{
    "symbol": "SMR",
    "chain_id": 1071,
    "contract": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "code": "smr",
    "type": 0,
    "decimal": 18,
    "public": 1
}
```

## GET v3/pools/all
```
Get all pools
```
### respose
```
See below response of /pools/{chain_id}/all
```

## GET v3/pools/{chain_id}/all
### respose
```json
[
  {
    "chain": 1071,
    "contract": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "token0": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "token1": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "fee_rate": "1000",
    "decimal":18
  }
]
```

## GET v3/pools/{chain_id}/{contract}
### respose
```json
{
    "chain": 1071,
    "contract": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "token0": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "token1": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "fee_rate": "1000",
    "decimal":18
}
```

## GET v3/pools/{chain_id}/overview
```
Get the overview of all pools by chain id
```
### respose
```json
[
    {
        "chainid":1071,
        "contract":"0x06A0464AAAc63335C8761b79e6B0C8F98232ab8F",
        "reserve0":"101513857227959068998",
        "reserve1":"4873095153106560344767",
        "volume24h0":"2105525145256533783",
        "volume24h1":"2105525145256533783",
        "utc0reserve0":"101513857227959068998",
        "utc0reserve1":"101513857227959068998",
        "ts":1685627419
    }
]
```

## GET v3/pools/{chain_id}/{contract}/overview
### respose
```json
{
    "chainid":1071,
    "contract":"0x06A0464AAAc63335C8761b79e6B0C8F98232ab8F",
    "reserve0":"101513857227959068998",
    "reserve1":"4873095153106560344767",
    "volume24h0":"2105525145256533783",
    "volume24h1":"2105525145256533783",
    "utc0reserve0":"101513857227959068998",
    "utc0reserve1":"101513857227959068998",
    "ts":1685627419
 }
```

## GET v3/pools/{chain_id}/{contract}/time-stats
### respose
```json
[
    {
        "id":19507,
        "contract":"",
        "reserve0":"69960295235747579314016",
        "reserve1":"577383448289",
        "vol01d":"14541618899316988008234",
        "vol11d":"14541618899316988008234",
        "vol07d":"14541618899316988008234",
        "vol17d":"14541618899316988008234"
    },
    {
        "id":19508,
        "contract":"",
        "reserve0":"69960295235747579314016",
        "reserve1":"577383448289",
        "vol01d":"14541618899316988008234",
        "vol11d":"14541618899316988008234",
        "vol07d":"14541618899316988008234",
        "vol17d":"14541618899316988008234"
    },
]
```

## GET v3/nfts/{user}/{collection}
```
Get all the nfts belong to user. collection if the nft's contract address
```
### respose
```json
[
    {
        "tokenid":"4",
        "collection":"0xEe610aE2b68b5549F231bf9152FFA2907a09ABC8",
        "user":"0x9A2c058A5020FAC6e316f11A0f1075DC930ac720",
        "pool":"0x99381366B094Cb94e88423A5cF604CFe536793dA",
        "token0":"0xc9f3a2C8a5C05FDbE086549de9DD9954ACA7BD22",
        "token1":"0xdcC4E969F081C3E967581Aa9175EF6F0a337Ae88",
        "fee":10000,
    }
]
```

# Scheduled Tasks
## 1. Record the state of each pools
Time schedule. Every 1 minute to record the state of each pool, including

* reserves
* volumes
* ts

## 2. Record the daily state of each pools
Time schedule. Every day at 00:00 UTC to record the daily state of each pool, including

* day,
* reserves,
* volumes1d,
* volumes7d.