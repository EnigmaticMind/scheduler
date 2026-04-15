# Build the Go scheduler binary
FROM golang:1.22-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /scheduler .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates

COPY --from=build /scheduler /usr/local/bin/scheduler

ENV PORT=8080

EXPOSE 8080

CMD ["/usr/local/bin/scheduler"]
