FROM golang:1.24 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

ADD *.go ./

RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o mutator .

# Use distroless as minimal base image to package the binary
FROM gcr.io/distroless/static:nonroot
LABEL source_repository="https://github.com/Facets-cloud/image-pull-secret-injector"
WORKDIR /
COPY --from=builder /workspace/mutator .
USER 65532:65532
ENTRYPOINT ["/mutator"]
