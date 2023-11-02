# api address
```
    https://dex.iotabee.com
```

# APIs for ibdex contract of V2 and V3

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

## GET /admin/pool/state
```
Change pool's state
```
### param
```
contract: 0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e
state: 0 or 1
address: user's address
ts: current timestamp
sign: sign for ts
```

## GET /admin/coin/public
```
Change coin's public
```
### param
```
contract: 0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e
public: 0 or 1
address: user's address
ts: current timestamp
sign: sign for ts
```

## GET /coins/all
```
Get all the tokens' data.
```
### Respose
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

## GET /coins/{contract}
```
Get a token by contract
```
### param
* contract: the address of token
### respose
```json
{
    "symbol": "SMR",
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
```json
[
  {
    "contract": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "version": 3,
    "token0": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "token1": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "fee_rate": "1000",
    "decimal":18
  }
]
```

## GET v3/pools/{contract}
### respose
```json
{
    "contract": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "version":3,
    "token0": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "token1": "0x4a8b2fbf8d3a0e3c3a8e571e3e6b9e8b5b6e5f5e",
    "fee_rate": "1000",
    "decimal":18,
    "state":0
}
```
state 0 is hide and 1 is public

## GET v3/pools/overview
```
Get the overview of all pools
```
### respose
```json
[
    {
        "contract":"0x06A0464AAAc63335C8761b79e6B0C8F98232ab8F",
        "reserve0":"101513857227959068998",
        "reserve1":"4873095153106560344767",
        "tick":0,
        "volume24h0":"2105525145256533783",
        "volume24h1":"2105525145256533783",
        "utc0reserve0":"101513857227959068998",
        "utc0reserve1":"101513857227959068998",
        "ts":1685627419
    }
]
```

## GET v3/pools/{contract}/overview
### respose
```json
{
    "contract":"0x06A0464AAAc63335C8761b79e6B0C8F98232ab8F",
    "reserve0":"101513857227959068998",
    "reserve1":"4873095153106560344767",
    "curr_tick":0,
    "volume24h0":"2105525145256533783",
    "volume24h1":"2105525145256533783",
    "utc0reserve0":"101513857227959068998",
    "utc0reserve1":"101513857227959068998",
    "utc0_tick":1,
    "ts":1685627419
 }
```

## GET v3/pools/{contract}/time-stats
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

