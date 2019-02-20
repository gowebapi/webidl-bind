#
# Used to verify that the public repository is buildable.
#
FROM golang:1.11-stretch
RUN go get github.com/gowebapi/webidl-bind
RUN go install github.com/gowebapi/webidl-bind
