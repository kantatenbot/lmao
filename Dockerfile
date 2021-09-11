FROM golang:alpine as builder

WORKDIR /build

ADD . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -o runtime ./cmd/mass-exec-runtime

RUN apk add git

WORKDIR /build/bin

RUN go get -v -u github.com/kantatenbot/bin/cmd/...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -o . github.com/kantatenbot/bin/cmd/...

FROM --platform=linux/amd64 ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && apt install -qy \
    locales \
    build-essential \
    libffi-dev \
    python3 \
    python3-dev \
    python3-pip

RUN apt install -qy \
    curl \
    dnsutils \
    file \
    git \
    jq

# install google cloud sdk (this is their "docker tip" from the docs)
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg  add - && apt-get update -y && apt-get install google-cloud-sdk -y

RUN pip install --no-cache-dir \
    awscli \
    boto3 \
    pydriller \
    requests

COPY --from=builder /build/runtime /opt/runtime/runtime

ADD bin /usr/bin/

ENTRYPOINT ["/opt/runtime/runtime"]
