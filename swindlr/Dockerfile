FROM golang:1.21 as builder

WORKDIR /app

# Copy go.mod and go.sum files to the container.
COPY go.mod go.mod ./

# Download dependencies.
RUN go mod download

# Copy the rest of the application source code.
COPY . .

# Build the application.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o swindlr .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/swindlr .
