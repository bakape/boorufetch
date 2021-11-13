package boorufetch

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type request struct {
	url string
	res chan<- response
}

type response struct {
	r   io.ReadCloser
	err error
}

type loadBalancer struct {
	queue        chan request
	scheme, host string
}

func newLoadBalancer(scheme, host string) *loadBalancer {
	l := &loadBalancer{
		queue:  make(chan request),
		scheme: scheme,
		host:   host,
	}
	for i := 0; i < FetcherCount; i++ {
		go func() {
			for {
				req := <-l.queue

				// Retry in case of throttling
			try:
				for i := 0; i < 3; i++ {
					r, err := http.Get(req.url)
					res := response{
						err: err,
					}
					if r != nil {
						switch r.StatusCode {
						case 200, 201:
							res.r = r.Body
						case 429, 500, 502, 503, 504:
							if r.Body != nil {
								r.Body.Close()
							}
							s := fmt.Sprintf(
								`GET %s returned status code %d; try %d/3`,
								req.url,
								r.StatusCode,
								i+1,
							)
							if i != 3 {
								s += `; retrying in 10s`
							}
							log.Printf(s)
							time.Sleep(time.Second * 10)
							continue try
						default:
							if r.Body != nil {
								r.Body.Close()
							}
							if res.err == nil {
								res.err = fmt.Errorf(
									`GET %s returned status code %d`,
									req.url,
									r.StatusCode,
								)
							}
						}
					}
					req.res <- res
					break try
				}
			}
		}()
	}
	return l
}

// Fetch resource by URL
func (l *loadBalancer) Fetch(url string) (io.ReadCloser, error) {
	ch := make(chan response)
	l.queue <- request{url, ch}
	r := <-ch
	return r.r, r.err
}
