# build a binary
FROM golang:1.22-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main ./cmd


# build a final small image with just a binary

FROM scratch

WORKDIR /app

COPY --from=builder /app/main ./

EXPOSE 80

CMD [ "./main", "serve" ]