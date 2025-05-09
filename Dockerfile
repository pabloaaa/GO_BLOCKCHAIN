FROM golang:1.19-alpine

WORKDIR /app

COPY . .

RUN go build -o server ./main.go # Ensure the correct file is built

EXPOSE 49990 49991 # Expose the correct ports

CMD ["./server"]
