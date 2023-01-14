package transfer

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

func TransferWemix() {
	// 블록체인 네트워크와 연결할 클라이언트를 생성하기 위한 rpc url 연결
	client, err := ethclient.Dial("https://api.test.wemix.com")
	if err != nil {
		fmt.Println("client error")
	}

	// metamask에서 뽑아낸 privatekey를 변환
	privateKey, err := crypto.HexToECDSA("")
	if err != nil {
		fmt.Println(err)
	}

	// privatekey로부터 publickey를 거쳐 자신의 address 변환
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("fail convert, publickey")
	}
	// 보낼 address 설정
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 현재 계정의 nonce를 가져옴. 다음 트랜잭션에서 사용할 nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		fmt.Println(err)
	}

	// 전송할 양, gasLimit, gasPrice 설정. 추천되는 gasPrice를 가져옴
	value := big.NewInt(700000000000000000)
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// 전송받을 상대방 address 설정
	toAddress := common.HexToAddress("")
	// 트랜잭션 생성
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// 트랜잭션 서명
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		fmt.Println(err)
	}

	// RLP 인코딩 전 트랜잭션 묶음. 현재는 1개의 트랜잭션
	ts := types.Transactions{signedTx}
	// RLP 인코딩
	rawTxBytes, _ := rlp.EncodeToBytes(ts[0])
	rawTxHex := hex.EncodeToString(rawTxBytes)
	rTxBytes, err := hex.DecodeString(rawTxHex)
	if err != nil {
		fmt.Println(err.Error())
	}

	// RLP 디코딩
	rlp.DecodeBytes(rTxBytes, &tx)
	// 트랜잭션 전송
	err = client.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println(err)
	}
	//출력된 tx.hash를 익스플로러에 조회 가능
	//예) 0x4788935cfa4a0f23807ba7d7b17a6304cc52795616889fdb9ebdb4498adf4a35
	fmt.Printf("tx sent: %s\n", tx.Hash().Hex())
}
