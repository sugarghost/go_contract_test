# go_contract_test

## 프로젝트 구조

```bash
config // 내부적으로 쓰이는 공통 값에 대한 config 저장
controller // 요청에 따른 로직 핸들링
logger // log 처리를 위한 구성
logs // 동작중 발생하는 사항에 대한 log 저장
model // 실제 이더리움 관련 처리 로직 담당
router // http 요청에 대한 controller 연결
keystore // 보안을 고려해 블록체인 개인키를 저장해 불러오기 위해 사용
contracts // 실제 계약 내용
```

### Route 구조

전체 경로는 `v1`으로 시작, 이후 `token`, `coin` 여부에 따라서 분기함

```go

	version1 := e.Group("v1", liteAuth())
	{
		token := version1.Group("token")
		{
			token.POST("/", p.ct.SendTokenByAddressController)

			token.GET("/symbol", p.ct.SearchTokenSymbolByTokenNameController)
			token.GET("/balance", p.ct.SearchTokenBalanceByAddressController)
			token.POST("/private", p.ct.SendTokenByAddressWithPrivateKeyController)
		}

		coin := version1.Group("coin")
		{
			coin.POST("/", p.ct.SendWemixCoinByAddressController)
			coin.POST("/private", p.ct.SendWemixCoinByAddressWithPrivateKeyController)
		}
	}

```

## keyStore

내부적으로 config에서 호출되 사용될 PrivateKey를 위해서 개인키를 keystore에 저장함

```go
    ks := keystore.NewKeyStore("./keystore", keystore.StandardScryptN, keystore.StandardScryptP)
    fmt.Println(ks)
    privateKey, _ := crypto.HexToECDSA("가짜 프라이빗 키 github에 올릴꺼")
    fmt.Println(privateKey)
    ks.ImportECDSA(privateKey, "가짜 비밀번호 github에 올릴꺼")
```

저장된 key는 `go run main.go` 실행시 `config`에 대한 설정에서 비밀번호를 입력받아 꺼내옴

```go
// config.go에 config 설정 과정 중 keystore 내용
    password := ""
    fmt.Print("keyStore 해금을 위한 Password : ")
    fmt.Scanf("%s", &password)
    account, err := keystore.DecryptKey(jsonBytes, password)
    if err == nil {
        pData := crypto.FromECDSA(account.PrivateKey)
        // Encode시 0x가 접두어로 붙기때문에 제거
        c.Contract.PrivateKey = hexutil.Encode(pData)[2:]
        fmt.Println(c)
        return c, nil
    }

```

## 처리로직

1. **`GET 이용` - 토큰 이름으로 토큰 심볼 조회**

```go
// controller 내용
func (p *Controller) SearchTokenSymbolByTokenNameController(c *gin.Context) {
	tokenName := c.Query("tokenName")
	symbol, err := p.md.SearchTokenSymbolByTokenNameModel(tokenName)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "symbol을 가져오지 못했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"symbol": symbol})
}

// model 내용


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
```

2. **`GET 이용` - 특정 주소가 소유한 토큰의 양 조회**

```go
// controller 내용
func (p *Controller) SearchTokenBalanceByAddressController(c *gin.Context) {

	address := c.GetHeader("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "address 정보가 유효하지 않습니다",
		})
		return
	}

	balance, err := p.md.SearchTokenBalanceByAddressModel(address)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "balance를 가져오지 못했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"balance": balance})
}

// model 내용
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

```

3. **`Post 이용` - 특정 주소에 지정한 양의 위믹스 코인을 전송**

```go
// controller 내용

func (p *Controller) SendWemixCoinByAddressController(c *gin.Context) {

	address := c.GetHeader("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "address 정보가 유효하지 않습니다",
		})
		return
	}
	err := p.md.SendWemixCoinByAddressModel(address, "")

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "전송에 실패했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"msg": "ok"})
}

// model 내용

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

```

4. **`Post 이용` - 다른 개인 키로, 특정 주소에 지정한 양의 위믹스 코인을 전송**

```go
// controller 내용

func (p *Controller) SendWemixCoinByAddressWithPrivateKeyController(c *gin.Context) {

	address := c.GetHeader("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "address 정보가 유효하지 않습니다",
		})
		return
	}
	privateKey := c.GetHeader("privateKey")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "privateKey 정보가 유효하지 않습니다",
		})
		return
	}
	err := p.md.SendWemixCoinByAddressModel(address, privateKey)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "전송에 실패했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"msg": "ok"})
}

// model 내용은 기존 SendWemixCoinByAddressModel 동일하며 개인키 입력 여부의 차이

```

5. **`Post 이용` - 특정 주소에 지정한 양의 토큰을 전송**

```go
// controller 내용
func (p *Controller) SendTokenByAddressController(c *gin.Context) {
	address := c.GetHeader("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "address 정보가 유효하지 않습니다",
		})
		return
	}
	err := p.md.SendTokenByAddressModel(address, "")

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "전송에 실패했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"msg": "ok"})
}
// model 내용

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
```

6. **`Post 이용` - 다른 개인 키로, 특정 주소에 지정한 양의 토큰을 전송**

```go
// controller 내용
func (p *Controller) SendTokenByAddressWithPrivateKeyController(c *gin.Context) {

	address := c.GetHeader("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "address 정보가 유효하지 않습니다",
		})
		return
	}
	privateKey := c.GetHeader("privateKey")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "privateKey 정보가 유효하지 않습니다",
		})
		return
	}
	err := p.md.SendTokenByAddressModel(address, privateKey)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "전송에 실패했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"msg": "ok"})
}

// model 내용은 기존 SendTokenByAddressModel과 동일하며 개인키 입력 여부의 차이
```
