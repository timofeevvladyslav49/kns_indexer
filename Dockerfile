FROM golang:1.25.4-alpine3.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /kns-indexer

FROM gcr.io/distroless/static-debian12

ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8

WORKDIR /

COPY --from=build-stage /kns-indexer /kns-indexer

USER nonroot:nonroot

ENTRYPOINT ["/kns-indexer"]