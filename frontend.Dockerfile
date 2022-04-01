FROM golang:latest

WORKDIR /usr/src/puzzles

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /usr/local/bin/frontend ./cmd/frontend/main.go

CMD ["frontend"]