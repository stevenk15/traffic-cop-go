FROM golang

# Add the package to the container
ADD . /go/src/github.com/stevenk15/traffic-cop-go

# Add the dependencies
RUN go get github.com/gocql/gocql
RUN go get gopkg.in/redis.v5

# Install the application in the container
RUN go install github.com/stevenk15/traffic-cop-go

# Run the application when the container starts
ENTRYPOINT /go/bin/traffic-cop-go

# Expose the container's port
EXPOSE 5000