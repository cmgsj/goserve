FROM golang:1.22.0 as build
WORKDIR /src
COPY . .
ENV GOOS="linux"
ENV GOARCH="amd64"
ENV CGO_ENABLED="0"
RUN make build

FROM scratch as runtime
COPY --from=build /src/bin /usr/local/bin
ENTRYPOINT ["goserve"]
