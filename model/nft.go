package model

import (
	"fmt"
	"strings"
)

func StoreNftToken(tokenId, collection, user, pool, token0, token1 string, fee int) error {
	_, err := db.Exec("INSERT INTO `nft`(`tokenid`,`collection`,`user`,`pool`,`token0`,`token1`,`fee`) VALUES(?,?,?,?,?,?,?)", tokenId, collection, user, pool, token0, token1, fee)
	if err != nil && strings.HasPrefix(err.Error(), "Error 1062") {
		return nil
	}
	return err
}

func DeleteNftToken(tokenId, collection string) error {
	_, err := db.Exec("delete from `nft` where `tokenid`=? and `collection`=?", tokenId, collection)
	return err
}

func TransferNftToOther(tokenId, collection, user string) error {
	_, err := db.Exec("update `nft` set `user`=? where `tokenid`=? and `collection`=?", user, tokenId, collection)
	return err
}

type NftToken struct {
	TokenId    string `json:"tokenid"`
	Collection string `json:"collection"`
	User       string `json:"user"`
	Pool       string `json:"pool"`
	Token0     string `json:"token0"`
	Token1     string `json:"token1"`
	Fee        int    `json:"fee"`
}

func GetNftTokens(user, collection string) ([]*NftToken, error) {
	rows, err := db.Query("select `tokenid`,`pool`,`token0`,`token1`,`fee` from `nft` where `user`=? and `collection`=?", user, collection)
	if err != nil {
		return nil, fmt.Errorf("get user's nft tokens from db error. %v", err)
	}
	ts := make([]*NftToken, 0)
	for rows.Next() {
		t := NftToken{Collection: collection, User: user}
		if err = rows.Scan(&t.TokenId, &t.Pool, &t.Token0, &t.Token1, &t.Fee); err != nil {
			return nil, fmt.Errorf("scan user's nft tokens from db error. %v", err)
		}
		ts = append(ts, &t)
	}
	return ts, nil
}
