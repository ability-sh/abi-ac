package ac

import (
	"fmt"
	"path/filepath"
	"runtime"

	unit "github.com/ability-sh/abi-ac/nginx-unit"
	"github.com/ability-sh/abi-lib/dynamic"
	"github.com/ability-sh/abi-lib/json"
	"github.com/ability-sh/abi-micro/micro"
)

type ACService struct {
	config    interface{} `json:"-"`
	name      string      `json:"-"`
	Unit      interface{} `json:"unit"`
	Container interface{} `json:"container"`
	Control   string      `json:"control"`
}

func newACService(name string, config interface{}) *ACService {
	return &ACService{name: name, config: config}
}

/**
* 服务名称
**/
func (s *ACService) Name() string {
	return s.name
}

/**
* 服务配置
**/
func (s *ACService) Config() interface{} {
	return s.config
}

func ac_CONFIG(item interface{}, appid string, ver string, ability string) {
	environment := dynamic.Get(item, "environment")
	if environment == nil {
		environment = map[string]interface{}{}
		dynamic.Set(item, "environment", environment)
	}
	dynamic.Set(environment, "AC_ENV", "unit")
	dynamic.Set(environment, "AC_APPID", appid)
	dynamic.Set(environment, "AC_VER", ver)
	dynamic.Set(environment, "AC_ABILITY", ability)
	config := dynamic.Get(item, "config")
	if config != nil {
		b, _ := json.Marshal(config)
		dynamic.Set(environment, "AC_CONFIG", string(b))
		dynamic.Set(item, "config", nil)
	}
}

/**
* 初始化服务
**/
func (s *ACService) OnInit(ctx micro.Context) error {

	ctx.Println(s.config)

	dynamic.SetValue(s, s.config)

	control := unit.NewControl(s.Control)

	var container Container = nil

	{
		t := dynamic.StringValue(dynamic.Get(s.Container, "type"), "")
		if t == "" || t == "default" {
			dir := dynamic.StringValue(dynamic.Get(s.Container, "dir"), "./apps")
			dir, _ = filepath.Abs(dir)
			container = NewAppContainer("abi-app", dir)
		}
	}

	if container == nil {
		return fmt.Errorf("not found container config")
	}

	var err error = nil
	var pkg Package = nil
	var ok bool = false

	pkgSet := map[string]Package{}

	dynamic.Each(dynamic.Get(s.Unit, "applications"), func(key interface{}, item interface{}) bool {

		appConfig := dynamic.Get(item, "app")

		if appConfig != nil {

			dynamic.Set(item, "app", nil)

			appid := dynamic.StringValue(dynamic.Get(appConfig, "appid"), "")
			ver := dynamic.StringValue(dynamic.Get(appConfig, "ver"), "")
			ability := dynamic.StringValue(dynamic.Get(appConfig, "ability"), "")

			key := fmt.Sprintf("%s_%s_%s", appid, ver, ability)

			pkg, ok = pkgSet[key]

			if !ok {

				pkg, err = container.GetPackage(ctx, appid, ver, ability)

				if err != nil {
					return false
				}

				pkgSet[key] = pkg

			}

			abilityConfig := dynamic.Get(pkg.Info(), ability)

			driver := dynamic.StringValue(dynamic.Get(abilityConfig, "driver"), "")
			root := dynamic.StringValue(dynamic.Get(abilityConfig, "root"), "")

			if driver == ":go" {
				ac_CONFIG(item, appid, ver, ability)
				dynamic.Set(item, "executable", fmt.Sprintf("bin/%s/%s/cloud", runtime.GOOS, runtime.GOARCH))
				dynamic.Set(item, "type", "external")
				dynamic.Set(item, "working_directory", pkg.Dir())
			} else if driver == ":node" {
				ac_CONFIG(item, appid, ver, ability)
				dynamic.Set(item, "executable", filepath.Join(root, "unit.js"))
				dynamic.Set(item, "type", "external")
				dynamic.Set(item, "working_directory", pkg.Dir())
			} else {
				err = fmt.Errorf("not found driver: %s", driver)
			}

		}
		return true
	})

	dynamic.Each(dynamic.Get(s.Unit, "routes"), func(key interface{}, item interface{}) bool {

		action := dynamic.Get(item, "action")

		if action != nil {

			share := dynamic.Get(action, "share")

			appid := dynamic.StringValue(dynamic.Get(share, "appid"), "")
			ver := dynamic.StringValue(dynamic.Get(share, "ver"), "")
			ability := dynamic.StringValue(dynamic.Get(share, "ability"), "")
			index := dynamic.StringValue(dynamic.Get(share, "index"), "")

			if appid != "" && ver != "" && ability != "" {

				key := fmt.Sprintf("%s_%s_%s", appid, ver, ability)

				pkg, ok = pkgSet[key]

				if !ok {

					pkg, err = container.GetPackage(ctx, appid, ver, ability)

					if err != nil {
						return false
					}

					pkgSet[key] = pkg

				}

				abilityConfig := dynamic.Get(pkg.Info(), ability)
				root := dynamic.StringValue(dynamic.Get(abilityConfig, "root"), "")

				if index != "" {
					dynamic.Set(action, "share", filepath.Join(pkg.Dir(), root, index))
				} else {
					dynamic.Set(action, "share", fmt.Sprintf("%s$uri", filepath.Join(pkg.Dir(), root)))
				}

			}

		}

		return true
	})

	ctx.Println(s.Unit)

	if err != nil {
		ctx.Println(err)
		return err
	}

	err = control.Put("http://localhost/config", s.Unit)

	if err != nil {
		ctx.Println(err)
		return err
	}

	return nil
}

/**
* 校验服务是否可用
**/
func (s *ACService) OnValid(ctx micro.Context) error {
	return nil
}

func (s *ACService) Recycle() {

}
