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

	containerId := os.Getenv("AC_ID")
	secret := os.Getenv("AC_SECRET")
	baseURL := os.Getenv("AC_BASE_URL")

	if baseURL == "" {
		baseURL = "https://ac.ability.sh"
		os.Setenv("AC_BASE_URL", baseURL)
	}

	flag.Parse()

	rd := bufio.NewReader(os.Stdin)

	var err error

	if containerId == "" {
		fmt.Printf("Please enter a Container ID: ")
		containerId, err = rd.ReadString('\n')
		containerId = strings.TrimSpace(containerId)
		os.Setenv("AC_ID", containerId)
		if err != nil {
			panic(err)
		}
	}

	if secret == "" {
		fmt.Printf("Please enter Secret: ")
		secret, err = rd.ReadString('\n')
		secret = strings.TrimSpace(secret)
		os.Setenv("AC_SECRET", secret)
		if err != nil {
			panic(err)
		}
	}

	p, err := runtime.NewAcPayload(baseURL, containerId, secret, runtime.NewPayload())

	if err != nil {
		panic(err)
	}

	defer p.Exit()

	select {}
}
