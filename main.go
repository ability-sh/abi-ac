package main

import (
	"log"
	"time"

	_ "github.com/ability-sh/abi-ac/ac"
	_ "github.com/ability-sh/abi-micro/crontab"
	_ "github.com/ability-sh/abi-micro/grpc"
	_ "github.com/ability-sh/abi-micro/http"
	_ "github.com/ability-sh/abi-micro/logger"
	"github.com/ability-sh/abi-micro/micro"
	_ "github.com/ability-sh/abi-micro/oss"
	_ "github.com/ability-sh/abi-micro/redis"
	"github.com/ability-sh/abi-micro/runtime"
)

func main() {

	tk := time.NewTicker(time.Second * 6)

	var err error = nil
	var payload micro.Payload = nil

	for {

		payload, err = runtime.NewFilePayload("./config.yaml", runtime.NewPayload())

		if err != nil {
			log.Println(err, "Wait 6 seconds to try again")
			<-tk.C
		} else {
			break
		}

	}

	tk.Stop()

	defer payload.Exit()

	select {}
}
