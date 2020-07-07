FROM golang:1.13-alpine AS build_finala

RUN apk add --update alpine-sdk git make && \
	git config --global http.https://gopkg.in.followRedirects true 

WORKDIR /app

COPY . .

RUN make build-linux && \
	mv /app/finala_linux /app/finala


FROM node:12.16-alpine AS build_ui

RUN apk add --update alpine-sdk make 

WORKDIR /app

COPY . .

RUN make build-ui

FROM alpine:3.9 
RUN apk add ca-certificates

COPY --from=build_ui /app/ui/build /ui/build
COPY --from=build_finala /app/finala /bin/finala

ENTRYPOINT ["/bin/finala"]