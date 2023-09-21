FROM golang:latest

ENV STORAGE_DIR=/mnt/data
RUN mkdir -p /mnt/data

WORKDIR /
COPY . .
RUN go build -ldflags "-s -w" -o cdn

CMD ["./cdn"]
EXPOSE 8080