package transfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/crypto/sha3"

	"go-contract/contracts" // 자신의 경로에 맞게 수정
)

func TransferCtxCoz() {
	// 블록체인 네트워크와 연결할 클라이언트를 생성하기 위한 rpc url 연결
	client, err := ethclient.Dial("https://api.test.wemix.com")
	if err != nil {
		fmt.Println("client error")
	}

	// 본인이 배포한 토큰 컨트랙트 어드레스
	tokenAddress := common.HexToAddress("0xe3236FEe84ffbcFA7955241CF0Bd0836169e075f")
	instance, err := contracts.NewContracts(tokenAddress, client)
	if err != nil {
		fmt.Println(err)
	}

	// 오너 어드레스
	address := common.HexToAddress("0x7C910BDA16C4774082DaAF7Ed88d94Ca7c45FcaF")
	bal, err := instance.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		fmt.Println(err)
	}

	// name 출력
	name, err := instance.Name(&bind.CallOpts{})
	if err != nil {
		fmt.Println(err)
	}

	// symbol 출력
	symbol, err := instance.Symbol(&bind.CallOpts{})
	if err != nil {
		fmt.Println(err)
	}

	// 사용되는 decimals 출력
	decimals, err := instance.Decimals(&bind.CallOpts{})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("balance: %s\n", bal)       // "balance: 999999999300000000000000000"
	fmt.Printf("name: %s\n", name)         // "name: Coz Token"
	fmt.Printf("symbol: %s\n", symbol)     // "symbol: Coz"
	fmt.Printf("decimals: %v\n", decimals) // "decimals: 18"

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
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 현재 계정의 nonce를 가져옴. 다음 트랜잭션에서 사용할 nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		fmt.Println(err)
	}

	// 전송할 양, gasLimit, gasPrice 설정. 추천되는 gasPrice를 가져옴
	value := big.NewInt(700000000000000000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// 보낼 주소
	toAddress := common.HexToAddress("0x5D86dE4B82091dBF1fd2c706d36ebC98E3d4d5Cd")

	// 컨트랙트 전송시 사용할 함수명
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID))

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d

	paddedAmount := common.LeftPadBytes(value.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000
	zvalue := big.NewInt(0)

	//컨트랙트 전송 정보 입력
	var pdata []byte
	pdata = append(pdata, methodID...)
	pdata = append(pdata, paddedAddress...)
	pdata = append(pdata, paddedAmount...)

	gasLimit := uint64(200000)
	fmt.Println(gasLimit)

	// 트랜잭션 생성
	tx := types.NewTransaction(nonce, tokenAddress, zvalue, gasLimit, gasPrice, pdata)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// 트랜잭션 서명
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		fmt.Println(err)
	}

	// 트랜잭션 전송
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fmt.Println(err)
	}

	//tx.hash를 이용해 전송결과를 확인
	//예)0x016430c748dad98865afb61038537f3ab8f504b56910769d328e7d857be7886a
	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
}
