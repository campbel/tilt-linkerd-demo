FROM golang:1.22-alpine

WORKDIR /app

# Copy go.mod and go.sum files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o /toxic .

WORKDIR /
CMD ["/toxic"]