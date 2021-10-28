FROM golang:1.16 as builder
WORKDIR /root
COPY . . 
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /root/p2p-server


FROM alpine:3.13
RUN apk --no-cache add curl
WORKDIR /root
COPY --from=builder /root/p2p-server ./
EXPOSE 8000
ENTRYPOINT ["/root/p2p-server"]  
