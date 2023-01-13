package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/naoina/toml"
)

type Config struct {
	Server struct {
		Port string
	}

	Contract struct {
		PrivateKey         string
		NetUrl             string
		TransactionHash    string
		TokenAddress       string
		ConstructorAddress string
	}

	KeyStore struct {
		Path string
	}
	Log struct {
		Level   string
		Fpath   string
		Msize   int
		Mage    int
		Mbackup int
	}
}

func NewConfig(fpath string) (*Config, error) {
	c := new(Config)
	file, err := os.Open(fpath)
	if err == nil {
		defer file.Close()
		//toml 파일 디코딩
		err := toml.NewDecoder(file).Decode(c)
		if err == nil {
			jsonBytes, err := ioutil.ReadFile(c.KeyStore.Path)
			if err == nil {
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
			}

		}
	}
	return nil, err
}
