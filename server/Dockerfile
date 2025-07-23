FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY ./go.* ./
RUN go mod download

COPY . ./

RUN go build -o /app './cmd'

FROM scratch

COPY --from=builder /app /app

ENTRYPOINT ["/app"]
CMD ["listen"]
