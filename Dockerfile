FROM golang:1.9.3-alpine3.7 AS build-env
RUN apk update && apk add build-base=0.5-r0 git=2.15.0-r1 bash=4.4.12-r2
ADD . /go/src/github.com/scaleci/scale
RUN cd /go/src/github.com/scaleci/scale && make build

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /go/src/github.com/scaleci/scale/build/scale /app/scale
ENTRYPOINT ./scale
