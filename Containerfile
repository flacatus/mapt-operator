FROM docker.io/library/golang:latest AS builder

ARG TARGETARCH
WORKDIR /workspace
COPY . .

# renovate: datasource=github-releases depName=pulumi/pulumi
ENV PULUMI_VERSION 3.178.0
ENV PULUMI_BASE_URL="https://github.com/pulumi/pulumi/releases/download/v${PULUMI_VERSION}/pulumi-v${PULUMI_VERSION}"
ENV PULUMI_URL="${PULUMI_BASE_URL}-linux-x64.tar.gz"

RUN unset VERSION \
    && go mod tidy \
    && go mod vendor \
    && make build \
    && if [ "$TARGETARCH" = "arm64" ]; then export PULUMI_URL="${PULUMI_BASE_URL}-linux-arm64.tar.gz"; fi \
    && curl -L ${PULUMI_URL} -o pulumicli.tar.gz \
    && tar -xzvf pulumicli.tar.gz

# ---------------------------------------------------

FROM registry.access.redhat.com/ubi9/ubi@sha256:861e833044a903f689ecfa404424494a7e387ab39cf7949c54843285d13a9774

ARG TARGETARCH
LABEL org.opencontainers.image.authors="Redhat Developer"

COPY --from=builder /workspace/bin/manager /workspace/pulumi/pulumi /usr/local/bin/

ENV PULUMI_CONFIG_PASSPHRASE "passphrase"
ENV AWS_SDK_LOAD_CONFIG=1 \
    AWS_CLI_VERSION=2.16.7 \
    AZ_CLI_VERSION=2.61.0 \
    ARCH_N=x86_64

# Pulumi plugins
ARG PULUMI_AWS_VERSION=v6.83.0
ARG PULUMI_AWSX_VERSION=v2.21.1
ARG PULUMI_AZURE_NATIVE_VERSION=v3.5.1
ARG PULUMI_COMMAND_VERSION=v1.1.0
ARG PULUMI_TLS_VERSION=v5.2.0
ARG PULUMI_RANDOM_VERSION=v4.18.2
ARG PULUMI_AWS_NATIVE_VERSION=v1.30.0

ENV PULUMI_HOME "/opt/cluster-info"
WORKDIR ${PULUMI_HOME}

RUN mkdir -p /opt/cluster-info \
    && if [ "$TARGETARCH" = "arm64" ]; then export ARCH_N=aarch64; fi \
    && export AWS_CLI_URL="https://awscli.amazonaws.com/awscli-exe-linux-${ARCH_N}-${AWS_CLI_VERSION}.zip" \
    && export AZ_CLI_RPM="https://packages.microsoft.com/rhel/9.0/prod/Packages/a/azure-cli-${AZ_CLI_VERSION}-1.el9.${ARCH_N}.rpm" \
    && curl ${AWS_CLI_URL} -o awscliv2.zip \
    && dnf install -y unzip \
    && unzip -qq awscliv2.zip \
    && ./aws/install \
    && curl -L ${AZ_CLI_RPM} -o azure-cli.rpm \
    && dnf install -y azure-cli.rpm \
    && rm -rf aws awscliv2.zip azure-cli.rpm \
    && dnf clean all \
    && rm -rf /var/cache/yum \
    && pulumi plugin install resource aws ${PULUMI_AWS_VERSION} \
    && pulumi plugin install resource azure-native ${PULUMI_AZURE_NATIVE_VERSION} \
    && pulumi plugin install resource command ${PULUMI_COMMAND_VERSION} \
    && pulumi plugin install resource tls ${PULUMI_TLS_VERSION} \
    && pulumi plugin install resource random ${PULUMI_RANDOM_VERSION} \
    && pulumi plugin install resource awsx ${PULUMI_AWSX_VERSION} \
    && pulumi plugin install resource aws-native ${PULUMI_AWS_NATIVE_VERSION}

RUN chmod +x /usr/local/bin/manager /usr/local/bin/pulumi \
    && mkdir -p /opt/cluster-info \
    && chmod 0777 /opt/cluster-info

ENTRYPOINT ["/usr/local/bin/manager"]
