FROM golang:1.18.0-alpine3.15 AS build

WORKDIR /app

RUN adduser -D scratchuser

COPY go.* ./
RUN go mod download

COPY src/*.go ./

RUN CGO_ENABLED=0 go build -o /syncer -ldflags="-s -w"

FROM scratch

WORKDIR /data

USER scratchuser

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /syncer /syncer

ENTRYPOINT ["/syncer"]

