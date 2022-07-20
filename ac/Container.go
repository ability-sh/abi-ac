package ac

import "github.com/ability-sh/abi-micro/micro"

type Container interface {
	GetPackage(ctx micro.Context, appid string, ver string, ability string) (Package, error)
}
