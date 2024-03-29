#build
FROM golang:1.16-alpine as builder
COPY ./ /pisec-brain
WORKDIR /pisec-brain/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -a -o brain .

#end user image
FROM alpine:3.11.3
COPY --from=builder pisec-brain/cmd/brain .
CMD ["./brain"]
