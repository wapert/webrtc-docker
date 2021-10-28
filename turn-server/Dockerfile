FROM golang:1.16 as builder
WORKDIR /root
COPY . . 
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /root/turn-server


FROM alpine:3.13
RUN apk --no-cache add curl
WORKDIR /root
COPY --from=builder /root/turn-server ./
EXPOSE 9000 19302
ENTRYPOINT ["/root/turn-server"]  
