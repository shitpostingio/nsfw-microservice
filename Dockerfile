# First stage: build the executable.
FROM ghcr.io/shitpostingio/tensorflow:latest AS builder

# It is important that these ARG's are defined after the FROM statement
ARG SSH_PRIV="nothing"
ARG SSH_PUB="nothing"
ARG GOSUMDB=off

# Create the user and group files that will be used in the running 
# container to run the process as an unprivileged user.
RUN mkdir /user && \
    echo 'nsfw:x:65534:65534:nsfw:/:' > /user/passwd && \
    echo 'nsfw:x:65534:' > /user/group

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/gitlab.com/shitposting/nsfw-microservice

# Import the code from the context.
COPY .  .

# Build the executable
RUN go install

# Final stage: the running container.
FROM ghcr.io/shitpostingio/tensorflow:latest

# Import the user and group files from the first stage.
COPY --from=builder /user/group /user/passwd /etc/

# Copy the built executable
COPY --from=builder /go/bin/nsfw-microservice /home/nsfw/nsfw-server

RUN chown -R nsfw /home/nsfw

# Expose the port
EXPOSE 10001

# Set the workdir
WORKDIR /home/nsfw

# Copy tensorflow model
COPY ./nsfw-lite ./nsfw-lite

# Perform any further action as an unprivileged user.
USER nsfw:nsfw

# Run the compiled binary.
CMD ["./nsfw-server"]
