package ac

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ability-sh/abi-lib/dynamic"
	"github.com/ability-sh/abi-lib/json"
)

type PackageInfo interface {
	Key() string
	Appid() string     // 应用ID
	Ver() string       // 应用版本
	Ability() []string // 应用包能力
	Info() interface{} // 应用包信息
}

type Package interface {
	PackageInfo
	Dir() string // 应用包目录
}

type packageInfo struct {
	appid   string
	ver     string
	ability []string
	info    interface{}
	key     string
}

func (p *packageInfo) Appid() string {
	return p.appid
}

func (p *packageInfo) Ver() string {
	return p.ver
}

func (p *packageInfo) Ability() []string {
	return p.ability
}

func (p *packageInfo) Info() interface{} {
	return p.info
}

func (p *packageInfo) Key() string {
	return p.key
}

func setPackageInfo(p *packageInfo, info interface{}) {
	p.appid = dynamic.StringValue(dynamic.Get(info, "appid"), "")
	p.ver = dynamic.StringValue(dynamic.Get(info, "ver"), "")
	ability := dynamic.StringValue(dynamic.Get(info, "ability"), "")
	if ability == "" {
		p.ability = []string{}
	} else {
		p.ability = strings.Split(ability, "|")
	}
	p.key = fmt.Sprintf("%s_%s", p.appid, p.ver)
	p.info = info
}

func NewPackageInfo(info interface{}) (PackageInfo, error) {
	p := &packageInfo{}
	setPackageInfo(p, info)
	if p.appid == "" || p.ver == "" || len(p.ability) == 0 {
		return nil, fmt.Errorf("app package info %s", info)
	}
	return p, nil
}

type pkg struct {
	info PackageInfo
	dir  string
}

func NewPackage(info PackageInfo, dir string) Package {
	return &pkg{info: info, dir: dir}
}

func (p *pkg) Appid() string {
	return p.info.Appid()
}

func (p *pkg) Ver() string {
	return p.info.Ver()
}

func (p *pkg) Ability() []string {
	return p.info.Ability()
}

func (p *pkg) Info() interface{} {
	return p.info.Info()
}

func (p *pkg) Key() string {
	return p.info.Key()
}

func (p *pkg) Dir() string {
	return p.dir
}

type dirPackage struct {
	packageInfo
	dir string
}

func GetInfoFile(p string) (interface{}, error) {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var info interface{} = nil
	err = json.Unmarshal(b, &info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func NewDirPackage(dir string) (Package, error) {

	p := dirPackage{}

	info, err := GetInfoFile(filepath.Join(dir, "app.json"))

	if err != nil {
		return nil, err
	}

	setPackageInfo(&p.packageInfo, info)

	if p.appid == "" || p.ver == "" || len(p.ability) == 0 {
		return nil, fmt.Errorf("app package error %s", dir)
	}

	p.dir = dir

	return &p, nil
}

func (p *dirPackage) Dir() string {
	return p.dir
}
