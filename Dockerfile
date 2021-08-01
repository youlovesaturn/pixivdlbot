FROM golang:alpine as build
WORKDIR /app
RUN apk update && apk add ca-certificates tzdata
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go/bin/app .


FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo/
COPY --from=build /go/bin/app /go/bin/app
ENTRYPOINT ["/go/bin/app"]