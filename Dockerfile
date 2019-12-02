FROM golang:1.12-alpine AS build_base

RUN apk add --update alpine-sdk git make && \
	git config --global http.https://gopkg.in.followRedirects true 

WORKDIR /app

CMD CGO_ENABLED=0 go test ./...

COPY . .

RUN go install -v ./...

FROM alpine:3.9 
RUN apk add ca-certificates

COPY --from=build_base /go/bin/finala /bin/finala

CMD ["finala"]