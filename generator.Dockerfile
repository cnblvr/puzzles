FROM golang:latest

WORKDIR /usr/src/puzzles

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /usr/local/bin/generator ./cmd/generator/main.go

CMD ["generator"]