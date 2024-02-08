FROM --platform=$BUILDPLATFORM golang:1.21
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o main ./main.go

# use debian as runtime image since muslc sucks
FROM --platform=$BUILDPLATFORM debian:bookworm-slim
COPY --from=0 /src/main /usr/local/bin/main
CMD ["main"]
