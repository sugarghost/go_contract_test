package main

import (
	"context"
	"flag"
	"fmt"
	conf "go-contract/config"
	ctl "go-contract/controller"
	log "go-contract/logger"
	md "go-contract/model"
	rt "go-contract/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

func main() {

	// 초기 keystore 생성을 위한 구문
	/*
		ks := keystore.NewKeyStore("./keystore", keystore.StandardScryptN, keystore.StandardScryptP)
		fmt.Println(ks)
		privateKey, _ := crypto.HexToECDSA("가짜 프라이빗 키 github에 올릴꺼")
		fmt.Println(privateKey)
		ks.ImportECDSA(privateKey, "가짜 비밀번호 github에 올릴꺼")
	*/

	var configFlag = flag.String("config", "./config/config.toml", "toml file to use for configuration")
	flag.Parse()

	if cf, err := conf.NewConfig(*configFlag); err != nil { // config 모듈 설정
		fmt.Printf("init config failed, err:%v\n", err)
		return
	} else if err := log.InitLogger(cf); err != nil { // logger 모듈 설정
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	} else if mod, err := md.NewModel(cf); err != nil { // model 모듈 설정
		fmt.Printf("NewModel Error: %v\n", err)
	} else if controller, err := ctl.NewCTL(mod); err != nil { //controller 모듈 설정
		fmt.Printf("NewCTL Error: %v\n", err)
	} else if rt, err := rt.NewRouter(controller); err != nil { //router 모듈 설정
		fmt.Printf("NewRouter Error: %v\n", err)
	} else {
		mapi := &http.Server{
			Addr:           cf.Server.Port,
			Handler:        rt.Idx(),
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		g.Go(func() error {
			return mapi.ListenAndServe()
		})

		stopSig := make(chan os.Signal)
		signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM)
		<-stopSig

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := mapi.Shutdown(ctx); err != nil {
			fmt.Println("Server Shutdown Error:", err)
		}

		select {
		case <-ctx.Done():
			fmt.Println("context done.")
		}
		fmt.Println("Server stop")
	}

	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}
}
