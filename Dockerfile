FROM golang:1.18.0-alpine3.15 AS build

RUN apk update && apk add openssh

WORKDIR /app

RUN adduser -D scratchuser

COPY go.* ./
RUN go mod download

COPY src/*.go ./

RUN CGO_ENABLED=0 go build -o /syncer -ldflags="-s -w"

USER scratchuser
RUN mkdir -p ~/.ssh && \
    chmod 700 ~/.ssh && \
    ssh-keyscan github.com > ~/.ssh/known_hosts && \
    chmod 600 ~/.ssh/known_hosts

FROM scratch

WORKDIR /data

USER scratchuser

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /home/scratchuser /home/scratchuser
COPY --from=build /syncer /syncer

ENTRYPOINT ["/syncer"]

