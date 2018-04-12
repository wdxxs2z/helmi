FROM golang:1.9-alpine as builder

# Install dependencies
RUN apk add --update --no-cache ca-certificates tar wget

# Build helmi
WORKDIR /go/src/github.com/wdxxs2z/helmi/

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o helmi .

# Copy helm artefacts
WORKDIR /app/
RUN cp /go/src/github.com/wdxxs2z/helmi/helmi .
RUN rm -r /go/src/

# Download kubectl 1.7.12
RUN wget -nv https://storage.googleapis.com/kubernetes-release/release/v1.7.12/bin/linux/amd64/kubectl && chmod 755 kubectl

# Download helm 2.8.2
RUN wget -nv -O- https://storage.googleapis.com/kubernetes-helm/helm-v2.8.2-linux-amd64.tar.gz | tar --strip-components=1 -zxf -

# Download dumb-init 1.2.1
RUN wget -nv -O /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.1/dumb-init_1.2.1_amd64 && chmod 755 /usr/local/bin/dumb-init

FROM alpine:3.7
RUN apk add --update --no-cache ca-certificates

WORKDIR /app/

COPY --from=builder /app/ .
COPY --from=builder /usr/local/bin/dumb-init /usr/local/bin/dumb-init

# Setup environment
ENV PATH "/app:${PATH}"

RUN addgroup -S helmi && \
    adduser -S -G helmi helmi && \
    chown -R helmi:helmi /app

USER helmi

ADD scripts/init_helm.sh /app/

RUN chmod +x /app/init_helm.sh

ENTRYPOINT ["/usr/local/bin/dumb-init", "--"]

CMD ["init_helm.sh"]