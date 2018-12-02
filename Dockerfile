FROM bakape/meguca
ENV GOPATH=/go
ENV PATH="${PATH}:/usr/local/go/bin:${GOPATH}/bin"
RUN mkdir -p /go/src/github.com/bakape/boorufetch
WORKDIR /go/src/github.com/bakape/boorufetch
COPY . .
RUN go get -v -t ./...
