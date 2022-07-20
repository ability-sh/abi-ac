package unit

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/ability-sh/abi-lib/json"
)

type Control struct {
	path string
}

func NewControl(path string) *Control {
	log.Println(path)
	return &Control{path: path}
}

func (C *Control) Do(req *http.Request) (*http.Response, error) {

	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: C.path, Net: "unix"})

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	err = req.Write(conn)

	if err != nil {
		return nil, err
	}

	return http.ReadResponse(bufio.NewReader(conn), req)
}

type result struct {
	Error   string `json:"error"`
	Success string `json:"success"`
	Detail  string `json:"detail"`
}

func (C *Control) Get(url string) (interface{}, error) {

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	res, err := C.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode == 200 {
		var rs interface{} = nil
		b, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &rs)
		if err != nil {
			return nil, err
		}
		return rs, nil
	} else {
		return nil, fmt.Errorf("%d: %s", res.StatusCode, res.Status)
	}
}

func (C *Control) Send(method, url string, config interface{}) error {

	body, err := json.Marshal(config)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))

	if err != nil {
		return err
	}

	res, err := C.Do(req)

	if err != nil {
		return err
	}

	rs := &result{}
	b, err := ioutil.ReadAll(res.Body)
	log.Println(string(b))
	res.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &rs)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("%s %s", rs.Error, rs.Detail)
	}
	return nil

}

func (C *Control) Put(url string, config interface{}) error {
	return C.Send("PUT", url, config)
}

func (C *Control) Post(url string, config interface{}) error {
	return C.Send("PUT", url, config)
}

func (C *Control) Delete(url string, config interface{}) error {
	return C.Send("DELETE", url, config)
}
