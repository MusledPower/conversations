FROM golang:1.25.8

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /migrate ./cmd/migrate
RUN go build -o /app ./cmd/app

ENV CONFIG_PATH=/app/config/local.yml

CMD ["./app"]