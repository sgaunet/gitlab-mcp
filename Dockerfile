FROM alpine:3.22.1 AS alpine

FROM scratch AS final
# Copy the ca-certificates.crt from the alpine image
COPY --from=alpine /etc/ssl/certs/ /etc/ssl/certs/
COPY "resources" /
WORKDIR /
COPY gitlab-mcp /gitlab-mcp
USER mcp
CMD ["/gitlab-mcp"]
