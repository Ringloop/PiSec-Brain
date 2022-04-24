FROM golang:1.16-alpine
COPY ./ /pisec-brain
RUN cd /pisec-brain && go build cmd/main.go
CMD ["/pisec-brain/main"]
