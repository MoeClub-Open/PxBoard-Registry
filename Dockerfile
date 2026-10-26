FROM golang:1.27-bookworm AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src/PxBoard-Registry
COPY . ./
COPY --from=core /go.mod /go.sum /src/PxBoard/
COPY --from=core /pkg /src/PxBoard/pkg/
COPY --from=official-plugins / /src/PxBoard-Plugins/
RUN go run ./cmd/publish -platform ${TARGETOS}-${TARGETARCH}
RUN CGO_ENABLED=0 go build -o /registry ./cmd/registry

FROM alpine:3.22
WORKDIR /registry
COPY --from=build /registry /usr/local/bin/registry
COPY --from=build /src/PxBoard-Registry/index.json ./index.json
COPY --from=build /src/PxBoard-Registry/packages ./packages
USER 10001
CMD ["registry", "-listen", "127.0.0.1:8091"]
