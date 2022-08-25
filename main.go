package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"flag"

	_ "github.com/ability-sh/abi-ac/ac"
	_ "github.com/ability-sh/abi-micro/crontab"
	_ "github.com/ability-sh/abi-micro/grpc"
	_ "github.com/ability-sh/abi-micro/http"
	_ "github.com/ability-sh/abi-micro/logger"
	_ "github.com/ability-sh/abi-micro/oss"
	_ "github.com/ability-sh/abi-micro/redis"
	"github.com/ability-sh/abi-micro/runtime"
)

func main() {

	containerId := flag.String("containerId", "", "containerId")
	secret := flag.String("secret", "", "secret")
	baseURL := flag.String("baseURL", "https://ac.ability.sh", "baseURL")

	flag.Parse()

	rd := bufio.NewReader(os.Stdin)

	var err error

	if *containerId == "" {
		fmt.Printf("Please enter a Container ID: ")
		*containerId, err = rd.ReadString('\n')
		*containerId = strings.TrimSpace(*containerId)
		if err != nil {
			panic(err)
		}
	}

	if *secret == "" {
		fmt.Printf("Please enter Secret: ")
		*secret, err = rd.ReadString('\n')
		*secret = strings.TrimSpace(*secret)
		if err != nil {
			panic(err)
		}
	}

	p, err := runtime.NewAcPayload(*baseURL, *containerId, *secret, runtime.NewPayload())

	if err != nil {
		panic(err)
	}

	defer p.Exit()

	select {}
}
