# Compile stage
FROM golang:1.16.3-alpine3.13 AS build-env

RUN apk add --update --no-cache --virtual .tmp-build-deps \
    gcc libc-dev linux-headers musl-dev zlib zlib-dev \
    libressl-dev libffi-dev
RUN apk add vips-dev

WORKDIR /src
COPY go.mod /src/
COPY go.sum /src/
COPY main.go /src/

COPY handlers/*.go /src/handlers/
COPY middleware/*.go /src/middleware/
COPY models/*.go /src/models/
COPY process/*.go /src/process/
COPY resources/*.go /src/resources/
COPY utils/*.go /src/utils/

RUN ls
RUN go build -o /server .

# Serve stage
FROM alpine:3.13

RUN apk add --update --no-cache --virtual .tmp-build-deps \
    gcc libc-dev linux-headers musl-dev zlib zlib-dev \
    libressl-dev libffi-dev
RUN apk add vips-dev exiftool

WORKDIR /
COPY static/ /static/
COPY templates/*.html /templates/
COPY --from=build-env /server /

EXPOSE 8080
CMD ["/server"]