package ac

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ability-sh/abi-lib/dynamic"
	"github.com/ability-sh/abi-lib/errors"
	"github.com/ability-sh/abi-micro/http"
	"github.com/ability-sh/abi-micro/micro"
)

type acContainer struct {
	baseURL string
	id      string
	secret  string
	dir     string
}

func NewACContainer(config interface{}) Container {
	dir, _ := filepath.Abs("./apps")
	return &acContainer{
		baseURL: dynamic.StringValue(dynamic.Get(config, "baseURL"), "https://ac.ability.sh"),
		id:      dynamic.StringValue(dynamic.Get(config, "id"), ""),
		secret:  dynamic.StringValue(dynamic.Get(config, "secret"), ""),
		dir:     dynamic.StringValue(dynamic.Get(config, "dir"), dir),
	}
}

func (c *acContainer) GetPackage(ctx micro.Context, appid string, ver string, ability string) (Package, error) {

	dir := filepath.Join(c.dir, ability, appid, ver)

	p, err := NewDirPackage(dir)

	if err == nil {
		return p, nil
	}

	xhr, err := http.GetHTTPService(ctx, "http")

	if err != nil {
		return nil, err
	}

	res, err := xhr.Request(ctx, "GET").
		SetURL(fmt.Sprintf("%s/store/container/app/get.json", c.baseURL), map[string]string{}).
		Send()

	if err != nil {
		return nil, err
	}

	rs, err := res.PraseBody()

	if err != nil {
		return nil, err
	}

	errno := int32(dynamic.IntValue(dynamic.Get(rs, "errno"), 0))

	if errno != 200 {
		return nil, errors.Errorf(errno, dynamic.StringValue(dynamic.Get(rs, "errmsg"), "Internal service error"))
	}

	info := dynamic.GetWithKeys(rs, []string{"data", "info"})
	info_md5 := dynamic.StringValue(dynamic.GetWithKeys(info, []string{ability, "md5"}), "")
	url := dynamic.StringValue(dynamic.GetWithKeys(rs, []string{"data", "url"}), "")

	if info_md5 == "" || url == "" {
		return nil, errors.Errorf(500, "App package not available")
	}

	err = os.MkdirAll(dir, os.ModePerm)

	if err != nil {
		return nil, err
	}

	zFile := filepath.Join(c.dir, fmt.Sprintf("%s-v%s-%s.zip", appid, ver, ability))

	os.Remove(zFile)

	fd, err := os.OpenFile(zFile, os.O_CREATE|os.O_WRONLY, os.ModePerm)

	if err != nil {
		return nil, err
	}

	mw := newMD5Writer(fd)

	res, err = xhr.Request(ctx, "GET").SetURL(url, nil).SetOutput(mw).Send()

	if err != nil {
		return nil, err
	}

	fd.Close()

	if res.Code() != 200 {
		return nil, fmt.Errorf("[%d] %s", res.Code(), string(res.Body()))
	}

	if info_md5 != hex.EncodeToString(mw.Sum(nil)) {
		return nil, errors.Errorf(500, "App package not available")
	}

	unz, err := zip.OpenReader(zFile)

	if err != nil {
		return nil, err
	}

	defer unz.Close()

	for _, f := range unz.File {
		fpath := filepath.Join(dir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {

			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return nil, err
			}

			inFile, err := f.Open()

			if err != nil {
				return nil, err
			}

			defer inFile.Close()

			os.Remove(fpath)

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE, os.ModePerm)

			if err != nil {
				return nil, err
			}

			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return nil, err
			}
		}
	}

	info_s, _ := json.MarshalIndent(info, "", "  ")

	err = ioutil.WriteFile(filepath.Join(dir, "app.json"), info_s, os.ModePerm)

	if err != nil {
		return nil, err
	}

	p, err = NewDirPackage(dir)

	if err != nil {
		return nil, err
	}

	return p, nil
}

type md5Writer struct {
	w io.Writer
	m hash.Hash
}

func newMD5Writer(w io.Writer) *md5Writer {
	return &md5Writer{w: w, m: md5.New()}
}

func (w *md5Writer) Write(p []byte) (n int, err error) {
	w.m.Write(p)
	return w.w.Write(p)
}

func (w *md5Writer) Sum(b []byte) []byte {
	return w.m.Sum(b)
}
