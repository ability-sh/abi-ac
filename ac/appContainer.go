package ac

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ability-sh/abi-lib/dynamic"
	"github.com/ability-sh/abi-lib/json"
	"github.com/ability-sh/abi-micro-app/pb"
	"github.com/ability-sh/abi-micro/grpc"
	"github.com/ability-sh/abi-micro/http"
	"github.com/ability-sh/abi-micro/micro"
)

type appContainer struct {
	name string
	dir  string
}

func NewAppContainer(name string, dir string) Container {
	return &appContainer{name: name, dir: dir}
}

func (c *appContainer) GetPackage(ctx micro.Context, appid string, ver string, ability string) (Package, error) {

	dir := filepath.Join(c.dir, ability, appid, ver)

	p, err := NewDirPackage(dir)

	if err == nil {
		return p, nil
	}

	conn, err := grpc.GetConn(ctx, c.name)

	if err != nil {
		return nil, err
	}

	xhr, err := http.GetHTTPService(ctx, "http")

	if err != nil {
		return nil, err
	}

	cli := pb.NewServiceClient(conn)

	ct := grpc.NewGRPCContext(ctx)

	var info interface{} = nil
	var info_s string = ""

	{
		rs, err := cli.VerGet(ct, &pb.VerGetTask{Appid: appid, Ver: ver})

		if err != nil {
			return nil, err
		}

		if rs.Errno != 200 {
			return nil, Errorf(int(rs.Errno), "%s", rs.Errmsg)
		}

		if rs.Data.Status != 1 {
			return nil, Errorf(500, "未找到应用包")
		}

		err = json.Unmarshal([]byte(rs.Data.Info), &info)

		if err != nil {
			return nil, err
		}
		info_s = rs.Data.Info
	}

	info_md5 := dynamic.StringValue(dynamic.Get(info, "md5"), "")

	if info_md5 == "" {
		return nil, Errorf(500, "错误的应用包")
	}

	rs, err := cli.VerGetURL(ct, &pb.VerGetURLTask{Appid: appid, Ver: ver, Ability: ability})

	if err != nil {
		return nil, err
	}

	if rs.Errno != 200 {
		return nil, Errorf(int(rs.Errno), "%s", rs.Errmsg)
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

	res, err := xhr.Request(ctx, "GET").SetURL(rs.Data, nil).SetOutput(mw).Send()

	if err != nil {
		return nil, err
	}

	fd.Close()

	if res.Code() != 200 {
		return nil, fmt.Errorf("[%d] %s", res.Code(), string(res.Body()))
	}

	if info_md5 != hex.EncodeToString(mw.Sum(nil)) {
		return nil, Errorf(500, "错误的应用包")
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

	err = ioutil.WriteFile(filepath.Join(dir, "app.json"), []byte(info_s), os.ModePerm)

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
