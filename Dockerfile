FROM node:12.16-alpine AS build_ui

RUN apk add --update alpine-sdk make 

WORKDIR /app

COPY . .

RUN make build-ui

FROM alpine:3.9 
RUN apk add ca-certificates curl wget jq

RUN DOWNLOAD_URL=$(curl -s https://api.github.com/repos/similarweb/finala/releases/latest \
  | jq -r '.assets[] | select(.browser_download_url | contains("Linux_i386")) | .browser_download_url') \
  && wget -qO- ${DOWNLOAD_URL} \
  | tar xz \
  && mv finala /bin/finala

COPY --from=build_ui /app/ui/build /ui/build

ENTRYPOINT ["/bin/finala"]