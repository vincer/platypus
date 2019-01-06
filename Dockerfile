FROM golang as builder

WORKDIR /go/src/github.com/vincer/platypus

COPY . .

RUN go get ./...

ENV CGO_ENABLED=0
RUN go install .

FROM scratch

COPY --from=builder /go/bin/platypus .

CMD ["./platypus"]

ENV PORT=5000
EXPOSE 5000
