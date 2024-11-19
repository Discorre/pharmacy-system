FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o pharmacy-system ./cmd/main.go

EXPOSE 8080

CMD ["./pharmacy-system"]
