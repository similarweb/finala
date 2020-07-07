FROM node:12.16-alpine AS build_ui

RUN apk add --update alpine-sdk make 

WORKDIR /app

COPY . .

RUN make build-ui

FROM alpine:3.9 
RUN apk add ca-certificates curl wget

RUN curl -s https://api.github.com/repos/similarweb/finala/releases/latest \
  | grep browser_download_url \
  | grep linux_386 \
  | cut -d '"' -f 4 \
  | wget -qi - && \
  tar -zxvf ./linux_386.tar.gz && \
  mv linux_386/finala /bin/finala

COPY --from=build_ui /app/ui/build /ui/build

ENTRYPOINT ["/bin/finala"]