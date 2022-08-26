package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

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

	{
		fs_containerId := flag.String("id", "", "Container ID")
		fs_secret := flag.String("secret", "", "Container Secret")
		fs_baseURL := flag.String("baseURL", "", "https://ac.ability.sh")

		flag.Parse()

		if *fs_containerId != "" {
			containerId = *fs_containerId
			os.Setenv("AC_ID", containerId)
		}
		if *fs_secret != "" {
			secret = *fs_secret
			os.Setenv("AC_SECRET", secret)
		}
		if *fs_baseURL != "" {
			baseURL = *fs_baseURL
			os.Setenv("AC_BASE_URL", baseURL)
		}
	}

	if baseURL == "" {
		baseURL = "https://ac.ability.sh"
		os.Setenv("AC_BASE_URL", baseURL)
	}

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
