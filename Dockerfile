FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/agent ./

FROM alpine:3.20
RUN adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=build /out/agent /app/agent
USER 10001
ENTRYPOINT ["/app/agent"]
