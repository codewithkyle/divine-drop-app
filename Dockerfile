# Step 1: Use an official Go image to build the Go application
FROM golang:1.22-alpine AS builder

# Step 2: Set the working directory inside the container
WORKDIR /server

# Step 3: Copy the Go modules and dependencies
COPY go.mod go.sum ./

# Step 4: Download necessary Go dependencies
RUN go mod download

# Step 5: Copy the entire Go project into the working directory
COPY . .

# Step 6: Build the Go application
RUN go build -ldflags="-s -w" -o divinedrop .

# Step 7: Create a small image to run the application
FROM alpine:latest

# Step 8: Set the working directory in the final image
WORKDIR /root/

# Step 9: Copy the Go binary from the builder stage
COPY --from=builder /server/divinedrop .

COPY . .

# Step 10: Expose the port that the app will run on
EXPOSE 3000

# Step 11: Command to run the Go app
CMD ["./divinedrop"]

