FROM public.ecr.aws/amazonlinux/amazonlinux:2023

ARG GO_VERSION=1.23.4
ARG TARGETARCH

RUN dnf install -y \
        git \
        tar \
        gzip \
        xz \
        unzip \
        jq \
        which \
        openssl \
        ca-certificates \
        shadow-utils \
        postgresql16 \
    && dnf clean all

RUN case "$(uname -m)" in \
      x86_64)  GOARCH=amd64 ;; \
      aarch64) GOARCH=arm64 ;; \
      *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;; \
    esac \
 && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz" \
      | tar -C /usr/local -xzf -

RUN case "$(uname -m)" in \
      x86_64)  AWSARCH=x86_64 ;; \
      aarch64) AWSARCH=aarch64 ;; \
    esac \
 && curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-${AWSARCH}.zip" -o /tmp/aws.zip \
 && unzip -q /tmp/aws.zip -d /tmp \
 && /tmp/aws/install \
 && rm -rf /tmp/aws /tmp/aws.zip

RUN dnf install -y dnf-plugins-core \
 && dnf config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo \
 && dnf install -y gh \
 && dnf clean all

RUN groupadd -g 1000 dev \
 && useradd -m -u 1000 -g 1000 -s /bin/bash dev

COPY dev/gh-app-token /usr/local/bin/gh-app-token
COPY dev/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/gh-app-token /usr/local/bin/entrypoint.sh

USER dev
WORKDIR /workspace

RUN curl -fsSL https://claude.ai/install.sh | bash

ENV PATH=/home/dev/.local/bin:/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["bash"]
