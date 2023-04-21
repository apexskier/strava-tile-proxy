FROM golang:1.20 as build
WORKDIR /app
ADD . /app
RUN go test ./...
RUN go build -o /binary

FROM gcr.io/distroless/base-debian10
COPY --from=build /binary /
CMD ["/binary"]
