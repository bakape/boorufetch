package boorufetch

import (
	"io"
	"net/http"
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
				r, err := http.Get(req.url)
				res := response{
					err: err,
				}
				if r != nil {
					res.r = r.Body
				}
				req.res <- res
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
