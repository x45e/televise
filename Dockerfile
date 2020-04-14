FROM golang:alpine as builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o televise ./cmd/televise
FROM scratch
EXPOSE 8080
COPY --from=builder /build/televise /app/
WORKDIR /app
ENTRYPOINT  ["/televise"]