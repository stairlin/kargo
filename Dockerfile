# Retrieve the `golang:alpine` image to provide us the
# necessary Golang tooling for building Go binaries.
# Here I retrieve the `alpine`-based just for the
# convenience of using a tiny image.
FROM golang:alpine as builder

ENV DEPV=v0.4.1
ENV BIN=kargo
ENV REPO=github.com/stairlin/$BIN

RUN mkdir -p /go/src/$REPO
WORKDIR /go/src/$REPO

# Install dep
RUN apk --no-cache add curl git && \
    curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/${DEPV}/dep-linux-amd64 && \
    chmod +x /usr/local/bin/dep

# Vendor dependencies
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

# Add remaining files
ADD . /go/src/$REPO

# 0.    Set some shell flags like `-e` to abort the
#       execution in case of any failure (useful if we
#       have many ';' commands) and also `-x` to print to
#       stderr each command already expanded.
# 1.    Get into the directory with the golang source code
# 2.    Perform the go build with some flags to make our
#       build produce a static binary (CGO_ENABLED=0 and
#       the `netgo` tag).
# 3.    copy the final binary to a suitable location that
#       is easy to reference in the next stage
RUN set -ex && \
  cd /go/src/${REPO} && \
  CGO_ENABLED=0 go build \
        -tags netgo \
        -v -a \
        -o ${BIN} \
        -ldflags '-extldflags "-static"' && \
  mv ./${BIN} /usr/bin/${BIN}

# Create the second stage with the most basic that we need - a
# busybox which contains some tiny utilities like `ls`, `cp`,
# etc. When we do this we'll end up dropping any previous
# stages (defined as `FROM <some_image> as <some_name>`)
# allowing us to start with a fat build image and end up with
# a very small runtime image. Another common option is using
# `alpine` so that the end image also has a package manager.
FROM busybox

# Retrieve the binary from the previous stage
COPY --from=builder /usr/bin/${BIN} /usr/local/bin/${BIN}

# Set the binary as the entrypoint of the container
ENTRYPOINT [ "kargo" ]