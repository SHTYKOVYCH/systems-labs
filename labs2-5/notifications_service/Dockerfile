FROM golang:latest

WORKDIR /app

COPY ./ ./

RUN go get .
RUN go mod download

RUN CGO_ENABLED=0 go build -o main ./main.go

CMD ./main
