FROM golang:1.12-alpine AS build_base

RUN apk add --update alpine-sdk git make && \
	git config --global http.https://gopkg.in.followRedirects true 

WORKDIR /app

COPY . .

RUN go install -v ./...

FROM alpine:3.9 
RUN apk add ca-certificates

COPY --from=build_base /app/ui/build /ui/build
COPY --from=build_base /app/config.yaml config.yaml
COPY --from=build_base /go/bin/finala /bin/finala

ENTRYPOINT ["finala"]