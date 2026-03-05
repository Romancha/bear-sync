FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/salmon-hub ./cmd/hub

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -h /data hub

USER hub
WORKDIR /data

COPY --from=build /bin/salmon-hub /usr/local/bin/salmon-hub

EXPOSE 7433

ENTRYPOINT ["salmon-hub"]
