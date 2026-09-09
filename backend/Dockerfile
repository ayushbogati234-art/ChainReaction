FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /chainreaction-api ./cmd/api && CGO_ENABLED=0 go build -o /chainreaction-worker ./cmd/worker
FROM alpine:3.21
COPY --from=build /chainreaction-api /chainreaction-api
COPY --from=build /chainreaction-worker /chainreaction-worker
EXPOSE 8080
