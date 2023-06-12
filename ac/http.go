package ac

import "fmt"

type HttpResponse struct {
	Code    int               `json:"code"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func (r *HttpResponse) String() string {
	return fmt.Sprintf("%v %v %v", r.Code, r.Headers, r.Body)
}

func (r *HttpResponse) Error() string {
	return fmt.Sprintf("%v %v %v", r.Code, r.Headers, r.Body)
}
