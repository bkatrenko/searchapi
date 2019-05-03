FROM golang:1.12.4 as builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

FROM scratch
COPY --from=builder /build/main  /app/
COPY --from=builder /build/certs /app/certs
COPY --from=builder /build/app.env /app/
WORKDIR /app
CMD ["./main"]