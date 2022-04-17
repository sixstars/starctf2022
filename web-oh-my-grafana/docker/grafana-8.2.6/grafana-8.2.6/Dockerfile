FROM golang:1.17.3-alpine3.14 as go-builder

RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
# RUN apk update
RUN apk add --no-cache gcc g++
# RUN apk add --no-cache make

RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY="https://goproxy.cn|direct"


WORKDIR $GOPATH/src/github.com/grafana/grafana

COPY go.mod go.sum embed.go ./
COPY cue cue
COPY cue.mod cue.mod
COPY packages/grafana-schema packages/grafana-schema
COPY public/app/plugins public/app/plugins
COPY pkg pkg
COPY build.go package.json ./

RUN go mod verify

RUN go run build.go build

# Final stage
FROM vulhub/grafana:8.2.6

COPY --from=go-builder /go/src/github.com/grafana/grafana/bin/*/grafana-server /go/src/github.com/grafana/grafana/bin/*/grafana-cli ./bin/

EXPOSE 3000

COPY ./packaging/docker/run.sh /run.sh

USER grafana
ENTRYPOINT [ "/run.sh" ]
