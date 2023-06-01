# api address
```
    https://dex.iotabee.com
```

# APIs for ibswap contract of V2

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

## GET /pools/all
```
Get all pools
```
### respose
```
See below response of /pools/{chain_id}/all
```

## GET /pools/{chain_id}/all
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

## GET /pools/{chain_id}/{contract}
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

## GET /pools/{chain_id}/overview
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

## GET /pools/{chain_id}/{contract}/overview
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

## GET /pools/{chain_id}/{contract}/time-stats
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