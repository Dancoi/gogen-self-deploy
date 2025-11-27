FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN apk add --no-cache git ca-certificates && go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /out/gogen-self-deploy ./

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/gogen-self-deploy /app/gogen-self-deploy
WORKDIR /app
ENTRYPOINT ["/app/gogen-self-deploy"]
