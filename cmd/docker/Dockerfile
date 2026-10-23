FROM golang:1.26

WORKDIR /app

COPY go.mod ./library/

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o myapp ./cmd/library/main.go

EXPOSE 8080

ENTRYPOINT ["/app/myapp"]
