package model

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	conf "go-contract/config"
	cont "go-contract/contracts"
	log "go-contract/logger"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

type Model struct {
	privateKey         string
	netUrl             string
	transactionHash    string
	tokenAddress       string
	constructorAddress string
}

func NewModel(cfg *conf.Config) (*Model, error) {
	r := &Model{}
	r.privateKey = cfg.Contract.PrivateKey
	r.netUrl = cfg.Contract.NetUrl
	r.transactionHash = cfg.Contract.TransactionHash
	r.tokenAddress = cfg.Contract.TokenAddress
	r.constructorAddress = cfg.Contract.ConstructorAddress

	return r, nil
}

func (p *Model) SearchTokenSymbolByTokenNameModel(tokenName string) (string, error) {

	// 블록체인 네트워크와 연결할 클라이언트를 생성하기 위한 rpc url 연결
	client, err := ethclient.Dial(p.netUrl)
	if err != nil {
		log.Error("client 에러", err.Error())
	}

	// 토큰 컨트랙트 어드레스
	tokenAddress := common.HexToAddress(p.tokenAddress)
	instance, err := cont.NewContractsCaller(tokenAddress, client)
	if err != nil {
		log.Error("NewContractsCaller 에러", err.Error())
	}

	contractTokenName, err := instance.Name(&bind.CallOpts{})
	if err != nil {
		log.Error("Token Name 조회 에러", err.Error())
		return "", err
	} else if contractTokenName != tokenName {
		log.Error("Token Name 불일치")
		return "", errors.New("token Name 불일치")
	}

	symbol, err := instance.Symbol(&bind.CallOpts{})
	if err != nil {
		log.Error("Symbol 조회 에러", err.Error())
		return "", err

	}

	return symbol, nil
}

func (p *Model) SearchTokenBalanceByAddressModel(address string) (*big.Int, error) {

	// 블록체인 네트워크와 연결할 클라이언트를 생성하기 위한 rpc url 연결
	client, err := ethclient.Dial(p.netUrl)
	if err != nil {
		log.Error("client 에러", err.Error())
	}

	// 토큰 컨트랙트 어드레스
	tokenAddress := common.HexToAddress(p.tokenAddress)
	instance, err := cont.NewContractsCaller(tokenAddress, client)
	if err != nil {
		log.Error("NewContractsCaller 에러", err.Error())
	}

	targetAddress := common.HexToAddress(address)
	balance, err := instance.BalanceOf(&bind.CallOpts{}, targetAddress)
	if err != nil {
		log.Error("balance 조회 에러", err.Error())
		return balance, err
	}

	return balance, nil
}

func (p *Model) SendTokenByAddressModel(targetAddress string, privateKeyParam string) error {

	// 블록체인 네트워크와 연결할 클라이언트를 생성하기 위한 rpc url 연결
	client, err := ethclient.Dial(p.netUrl)
	if err != nil {
		log.Error("client 에러", err.Error())
		return err
	}

	// 토큰 컨트랙트 어드레스
	tokenAddress := common.HexToAddress(p.tokenAddress)

	if privateKeyParam == "" {
		privateKeyParam = p.privateKey
	}
	// 기본키 지정
	privateKey, err := crypto.HexToECDSA(privateKeyParam)
	if err != nil {
		log.Error("HexToECDSA 에러", err.Error())
		return err
	}

	// privatekey로부터 publickey를 거쳐 자신의 address 변환
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Error("fail convert, publickey")
		return errors.New("fail convert")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 현재 계정의 nonce를 가져옴. 다음 트랜잭션에서 사용할 nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Error("PendingNonceAt 에러", err.Error())
		return err
	}

	// 전송할 양, gasLimit, gasPrice 설정. 추천되는 gasPrice를 가져옴
	value := big.NewInt(700000000000000000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error("SuggestGasPrice 에러", err.Error())
		return err
	}

	// 보낼 주소
	toAddress := common.HexToAddress(targetAddress)

	// 컨트랙트 전송시 사용할 함수명
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID))

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(value.Bytes(), 32)
	zvalue := big.NewInt(0)

	//컨트랙트 전송 정보 입력
	var pdata []byte
	pdata = append(pdata, methodID...)
	pdata = append(pdata, paddedAddress...)
	pdata = append(pdata, paddedAmount...)

	gasLimit := uint64(200000)

	// 트랜잭션 생성
	tx := types.NewTransaction(nonce, tokenAddress, zvalue, gasLimit, gasPrice, pdata)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Error("트랜잭션 생성 에러", err.Error())
		return err
	}

	// 트랜잭션 서명
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Error("트랜잭션 서명 에러", err.Error())
		return err
	}

	// 트랜잭션 전송
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Error("트랜잭션 전송 에러", err.Error())
		return err
	}

	//tx.hash를 이용해 전송결과를 확인
	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
	return nil
}

func (p *Model) SendWemixCoinByAddressModel(targetAddress string, privateKeyParam string) error {

	// 블록체인 네트워크와 연결할 클라이언트를 생성하기 위한 rpc url 연결
	client, err := ethclient.Dial(p.netUrl)
	if err != nil {
		log.Error("client 에러", err.Error())
		return err
	}

	if privateKeyParam == "" {
		privateKeyParam = p.privateKey
	}
	// 기본키 지정
	privateKey, err := crypto.HexToECDSA(privateKeyParam)
	if err != nil {
		log.Error("HexToECDSA 에러", err.Error())
		return err
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
		log.Error("PendingNonceAt 에러", err.Error())
		return err
	}

	// 전송할 양, gasLimit, gasPrice 설정. 추천되는 gasPrice를 가져옴
	value := big.NewInt(700000000000000000)
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error("SuggestGasPrice 에러", err.Error())
		return err
	}

	// 보낼 주소
	toAddress := common.HexToAddress(targetAddress)

	// 트랜잭션 생성
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Error("트랜잭션 생성 에러", err.Error())
		return err
	}

	// 트랜잭션 서명
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Error("트랜잭션 서명 에러", err.Error())
		return err
	}

	// RLP 인코딩 전 트랜잭션 묶음. 현재는 1개의 트랜잭션
	ts := types.Transactions{signedTx}
	// RLP 인코딩
	rawTxBytes, _ := rlp.EncodeToBytes(ts[0])
	rawTxHex := hex.EncodeToString(rawTxBytes)
	rTxBytes, err := hex.DecodeString(rawTxHex)
	if err != nil {
		log.Error("RLP 인코딩 에러", err.Error())
		return err
	}

	// RLP 디코딩
	rlp.DecodeBytes(rTxBytes, &tx)

	// 트랜잭션 전송
	err = client.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Error("트랜잭션 전송 에러", err.Error())
		return err
	}

	//tx.hash를 이용해 전송결과를 확인
	fmt.Printf("tx sent: %s\n", tx.Hash().Hex())
	return nil
}
