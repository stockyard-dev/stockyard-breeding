FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go mod download && CGO_ENABLED=0 go build -o breeding ./cmd/breeding/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/breeding .
ENV PORT=9803 DATA_DIR=/data
EXPOSE 9803
CMD ["./breeding"]
