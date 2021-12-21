FROM debian:buster-slim
RUN apt update && \
    apt install -y curl
# Work inside the /tmp directory
WORKDIR /tmp
RUN curl https://storage.googleapis.com/golang/go1.16.2.linux-amd64.tar.gz -o go.tar.gz && \
    tar -zxf go.tar.gz && \
    rm -rf go.tar.gz && \
    mv go /go
ENV GOPATH /go
ENV PATH $PATH:/go/bin:$GOPATH/bin
# If you enable this, then gcc is needed to debug your app
ENV CGO_ENABLED 0

RUN mkdir -p /go/src/github.com/aquasecurity/trivy
WORKDIR /go/src/github.com/aquasecurity/trivy
COPY . .
ENV GOARCH amd64
ENV GOOS linux
RUN go get -t ./...
RUN go build -o trivy github.com/aquasecurity/trivy/cmd/trivy
