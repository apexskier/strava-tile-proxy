FROM golang:1.24 AS build
WORKDIR /app
ADD . /app
RUN go test ./...
RUN go build -o /binary

FROM gcr.io/distroless/base-debian12
COPY --from=build /binary /
CMD ["/binary"]
