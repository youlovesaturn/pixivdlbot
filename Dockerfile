FROM golang:alpine as build
RUN apk update && apk add ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go/bin/app .


FROM scratch
WORKDIR /app
COPY --from=build /go/bin/app /go/bin/app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo/
ENTRYPOINT ["/go/bin/app"]