package ac

import (
	"github.com/ability-sh/abi-micro/micro"
)

func init() {
	micro.Reg("abi-ac", func(name string, config interface{}) (micro.Service, error) {
		return newACService(name, config), nil
	})
}
