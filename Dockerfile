# Stage 1: This stage is no longer for building Go, but for preparing the final image.
# We expect the binary to be pre-built by CI and available in the build context.
FROM alpine:latest

WORKDIR /app

# Arguments to specify the target OS and architecture of the pre-built binary
ARG TARGETOS
ARG TARGETARCH

# Copy the pre-built binary from the build context (e.g., from 'dist/linux-amd64/ddns-dnspod')
# The path in the COPY command must match where CI places the binaries.
COPY dist/${TARGETOS}-${TARGETARCH}/ddns-dnspod /app/ddns-dnspod

# Ensure the binary is executable
RUN chmod +x /app/ddns-dnspod

# Declare expected environment variables.
# For sensitive variables like secrets, do NOT provide actual default values here.
# These ENV lines primarily serve as documentation within the Dockerfile and
# can set non-sensitive defaults if applicable.
# Users will need to provide these at runtime via `docker run -e VAR=value ...`
ENV DNSPOD_SECRET_ID=""
ENV DNSPOD_SECRET_KEY=""
ENV DNSPOD_DOMAIN=""
ENV DNSPOD_RECORDID_IPV4=""
ENV DNSPOD_SUBDOMAIN_IPV4="@"
ENV DNSPOD_RECORDID_IPV6=""
ENV DNSPOD_SUBDOMAIN_IPV6="@"

# Application will look for config.toml in the same directory as the executable,
# or rely on environment variables.
# It's recommended to mount config.toml or use environment variables for configuration.

# The application logs to stdout/stderr when run directly (which is how Docker runs it),
# and also attempts to write to ddns-server.log in its working directory.
# Ensure the working directory is writable if file logging is desired within the container.
# For production, rely on Docker's logging mechanisms for stdout/stderr.

# Run the application.
ENTRYPOINT ["/app/ddns-dnspod"]

# Default command arguments (e.g., path to config if not using env vars and not in /app)
# CMD ["-c", "/app/config.toml"]
