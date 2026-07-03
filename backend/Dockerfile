# NOTE: this image has not been built in the environment that authored it (no Docker
# available there). go.sum does not exist yet, so `go mod tidy` runs inside the build stage
# to resolve dependencies from scratch — regenerate and commit go.sum locally once you have
# a Go toolchain, then switch this back to `go mod download` for reproducible/cached builds.

FROM golang:1.23 AS build
WORKDIR /src
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
