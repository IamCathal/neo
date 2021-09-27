# builder image
FROM golang:1.15.7-alpine as builder
RUN mkdir /build
ADD *.go /build/
ADD go.mod /build/
ADD go.sum /build/
ADD .env /build/
WORKDIR /build
RUN go install
RUN CGO_ENABLED=0 GOOS=linux go build -a -o frontend .


# generate clean, final image for end users
FROM alpine:3.13.6
COPY --from=builder /build/frontend .
COPY --from=builder /build/.env .
# executable
ENTRYPOINT [ "./frontend" ]
# arguments that can be overridden
# CMD [ "3", "300" ]