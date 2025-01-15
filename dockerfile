# Start from the official Go image
FROM golang:1.23-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code into the container
COPY . .

# Install dependencies (this step can be skipped if there are no dependencies)
RUN go mod tidy

# Build the Go application from cmd/main.go
RUN go build -o main ./cmd

# Expose the port the app will run on
EXPOSE 5000

# Run the compiled Go binary
CMD ["./main"]
