FROM golang:1.18.0-stretch@sha256:6e60e855593ac027bafd503eba1a4d29a5551901092a60ebd199dbb23ef691b2 AS backend-builder
WORKDIR /go/src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
ARG VERSION=unknown
RUN go install -mod=readonly -ldflags "-X main.version=$VERSION" .
RUN go test ./...


FROM node@sha256:58fef6ad5c242f1cdd9259e15fa540c40dfc8292704c9d37baa23420a7cfa910 AS frontend-builder
WORKDIR /tmp/front
COPY ./front/package*.json ./
RUN npm install
COPY ./front .
RUN ./node_modules/.bin/vue-cli-service build --dest=dist src/main.js


FROM debian@sha256:3d3251ecfd190284f76c271e6c9dd19bb4d111f10bb9028ad8b8ac3d3b4914ea

RUN apt update && apt install -y ca-certificates && apt clean

WORKDIR /opt/coroot

COPY --from=backend-builder /go/bin/coroot /opt/coroot/coroot
COPY --from=frontend-builder /tmp/front/dist /opt/coroot/static

VOLUME /data
EXPOSE 8080

ENTRYPOINT ["/opt/coroot/coroot"]
