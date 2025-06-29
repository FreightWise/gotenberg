# ARG instructions do not create additional layers. Instead, next layers will
# concatenate them. Also, we have to repeat ARG instructions in each build
# stage that uses them.
ARG GOLANG_VERSION=1.24

# ----------------------------------------------
# pdfcpu binary build stage
# ----------------------------------------------
# Note: this stage is required as pdfcpu does not release an armhf variant by
# default.
FROM golang:$GOLANG_VERSION AS pdfcpu-binary-stage

# See https://github.com/pdfcpu/pdfcpu/releases.
ARG PDFCPU_VERSION=v0.8.1 
ENV CGO_ENABLED=0

# Define the working directory outside of $GOPATH (we're using go modules).
WORKDIR /home

RUN curl -Ls "https://github.com/pdfcpu/pdfcpu/archive/refs/tags/$PDFCPU_VERSION.tar.gz" -o pdfcpu.tar.gz &&\
    tar --strip-components=1 -xvzf pdfcpu.tar.gz

# Install module dependencies.
RUN go mod download &&\
    go mod verify

RUN go build -o pdfcpu -ldflags "-s -w -X 'main.version=$PDFCPU_VERSION' -X 'github.com/pdfcpu/pdfcpu/pkg/pdfcpu.VersionStr=$PDFCPU_VERSION' -X main.builtBy=gotenberg" ./cmd/pdfcpu &&\
    # Verify installation.
    ./pdfcpu version

# ----------------------------------------------
# Gotenberg binary build stage
# ----------------------------------------------
FROM golang:$GOLANG_VERSION AS gotenberg-binary-stage

ARG GOTENBERG_VERSION=snapshot
ENV CGO_ENABLED=0

# Define the working directory outside of $GOPATH (we're using go modules).
WORKDIR /home

# Install module dependencies.
COPY go.mod go.sum ./

RUN go mod download &&\
    go mod verify

# Copy the source code.
COPY cmd ./cmd
COPY pkg ./pkg

RUN go build -o gotenberg -ldflags "-s -w -X 'github.com/gotenberg/gotenberg/v8/cmd.Version=$GOTENBERG_VERSION'" cmd/gotenberg/main.go

# ----------------------------------------------
# Compress Golang binaries stage
# ----------------------------------------------
FROM debian:12-slim AS compress-go-binaries-stage

RUN \
    echo "deb http://deb.debian.org/debian bookworm-backports main" >> /etc/apt/sources.list &&\
    apt-get update -qq &&\
    apt-get upgrade -yqq &&\
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends -t bookworm-backports upx-ucl

COPY --from=pdfcpu-binary-stage /home/pdfcpu /home/pdfcpu
COPY --from=gotenberg-binary-stage /home/gotenberg /home/gotenberg

RUN \
    upx-ucl --best /home/pdfcpu &&\
    upx-ucl --best /home/gotenberg

# ----------------------------------------------
# Custom JRE stage
# Credits: https://github.com/jodconverter/docker-image-jodconverter-runtime
# ----------------------------------------------
FROM debian:12-slim AS custom-jre-stage

RUN \
    apt-get update -qq &&\
    apt-get upgrade -yqq &&\
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends default-jdk-headless binutils

# Note: jdeps helps finding which modules a JAR requires.
# Currently only for PDFtk, as we don't rely on LibreOffice UNO Java SDK.
ENV JAVA_MODULES=java.base,java.desktop,java.naming,java.sql

RUN jlink \
    --add-modules $JAVA_MODULES \
    --strip-debug \
    --no-man-pages \
    --no-header-files \
    --compress=2 \
    --output /custom-jre

# ----------------------------------------------
# Base image stage
# ----------------------------------------------
FROM debian:12-slim AS base-image-stage

COPY --from=custom-jre-stage /custom-jre /opt/java

ENV PATH="/opt/java/bin:${PATH}"

# ----------------------------------------------
# Final stage
# ----------------------------------------------
FROM base-image-stage

ARG GOTENBERG_VERSION=snapshot
ARG GOTENBERG_USER_GID=1001
ARG GOTENBERG_USER_UID=1001
# See https://github.com/googlefonts/noto-emoji/releases.
ARG NOTO_COLOR_EMOJI_VERSION=v2.047
# See https://gitlab.com/pdftk-java/pdftk/-/releases - Binary package.
ARG PDFTK_VERSION=v3.3.3

LABEL org.opencontainers.image.title="Gotenberg" \
    org.opencontainers.image.description="A containerized API for seamless PDF conversion." \
    org.opencontainers.image.version="$GOTENBERG_VERSION" \
    org.opencontainers.image.authors="Julien Neuhart <neuhart.julien@gmail.com>" \
    org.opencontainers.image.documentation="https://gotenberg.dev" \
    org.opencontainers.image.source="https://github.com/gotenberg/gotenberg"

RUN \
    # Create a non-root user.
    # All processes in the Docker container will run with this dedicated user.
    groupadd --gid "$GOTENBERG_USER_GID" gotenberg &&\
    useradd --uid "$GOTENBERG_USER_UID" --gid gotenberg --shell /bin/bash --home /home/gotenberg --no-create-home gotenberg &&\
    mkdir /home/gotenberg &&\
    chown gotenberg: /home/gotenberg

RUN \
    # Install system dependencies required for the next instructions or debugging.
    # Note: tini is a helper for reaping zombie processes.
    apt-get update -qq &&\
    apt-get upgrade -yqq &&\
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends curl gnupg=2.2.40-1.1 tini=0.19.0-1 python3=3.11.2-1+b1 &&\
    # Cleanup.
    # Note: the Debian image does automatically a clean after each install thanks to a hook.
    # Therefore, there is no need for apt-get clean.
    # See https://stackoverflow.com/a/24417119/3248473.
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN \
    # Install fonts.
    # Credits:
    # https://github.com/arachnys/athenapdf/blob/master/cli/Dockerfile.
    # https://help.accusoft.com/PrizmDoc/v12.1/HTML/Installing_Asian_Fonts_on_Ubuntu_and_Debian.html.
    curl -o ./ttf-mscorefonts-installer_3.8.1_all.deb http://httpredir.debian.org/debian/pool/contrib/m/msttcorefonts/ttf-mscorefonts-installer_3.8.1_all.deb &&\
    apt-get update -qq &&\
    apt-get upgrade -yqq &&\
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends \
    ./ttf-mscorefonts-installer_3.8.1_all.deb \
    culmus=0.133-1 \
    fonts-beng=2:1.3 \
    fonts-hosny-amiri=0.113-1 \
    fonts-lklug-sinhala=0.6-4 \
    fonts-lohit-guru=2.91.2-3 \
    fonts-lohit-knda=2.5.4-3 \
    fonts-samyak-gujr=1.2.2-6 \
    fonts-samyak-mlym=1.2.2-6 \
    fonts-samyak-taml=1.2.2-6 \
    fonts-sarai=1.0-3 \
    fonts-sil-abyssinica=2.100-3 \
    fonts-sil-padauk=5.000-3 \
    fonts-telu=2:1.3 \
    fonts-thai-tlwg=1:0.7.3-1 \
    ttf-wqy-zenhei \
    fonts-arphic-ukai=0.2.20080216.2-5 \
    fonts-arphic-uming=0.2.20080216.2-11 \
    fonts-ipafont-mincho=00303-23 \
    fonts-ipafont-gothic=00303-23 \
    fonts-unfonts-core=1:1.0.2-080608-18 \
    # LibreOffice recommends.
    fonts-crosextra-caladea=20200211-1 \
    fonts-crosextra-carlito=20220224-1 \
    fonts-dejavu=2.37-6 \
    fonts-liberation=1:1.07.4-11 \
    fonts-liberation2=2.1.5-1 \
    fonts-linuxlibertine=5.3.0-6 \
    fonts-noto-cjk=1:20220127+repack1-1 \
    fonts-noto-core=20201225-1 \
    fonts-noto-mono=20201225-1 \
    fonts-noto-ui-core=20201225-1 \
    fonts-sil-gentium=20081126:1.03-4 \
    fonts-sil-gentium-basic=1.102-1.1 &&\
    rm -f ./ttf-mscorefonts-installer_3.8.1_all.deb &&\
    # Add Color and Black-and-White Noto emoji font.
    # Credits:
    # https://github.com/gotenberg/gotenberg/pull/325.
    # https://github.com/googlefonts/noto-emoji.
    curl -Ls "https://github.com/googlefonts/noto-emoji/raw/$NOTO_COLOR_EMOJI_VERSION/fonts/NotoColorEmoji.ttf" -o /usr/local/share/fonts/NotoColorEmoji.ttf &&\
    # Cleanup.
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN \
    # Install either Google Chrome stable on amd64 architecture or
    # Chromium on other architectures.
    # See https://github.com/gotenberg/gotenberg/issues/328.
    /bin/bash -c \
    'set -e &&\
    if [[ "$(dpkg --print-architecture)" == "amd64" ]]; then \
    curl https://dl.google.com/linux/linux_signing_key.pub | apt-key add - &&\
    echo "deb http://dl.google.com/linux/chrome/deb/ stable main" | tee /etc/apt/sources.list.d/google-chrome.list &&\
    apt-get update -qq &&\
    apt-get upgrade -yqq &&\
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends --allow-unauthenticated google-chrome-stable &&\
    mv /usr/bin/google-chrome-stable /usr/bin/chromium; \
    else \
    apt-get update -qq &&\
    apt-get upgrade -yqq &&\
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends chromium; \
    fi' &&\
    # Verify installation.
    chromium --version &&\
    # Cleanup.
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Set default characterset encoding to UTF-8.
# See:
# https://github.com/gotenberg/gotenberg/issues/104
# https://github.com/gotenberg/gotenberg/issues/730
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

RUN \
    # Install LibreOffice & unoconverter.
    echo "deb http://deb.debian.org/debian bookworm-backports main" >> /etc/apt/sources.list &&\
    apt-get update -qq &&\
    apt-get upgrade -yqq &&\
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends -t bookworm-backports libreoffice &&\
    curl -Ls https://raw.githubusercontent.com/gotenberg/unoconverter/v0.1.1/unoconv -o /usr/bin/unoconverter &&\
    chmod +x /usr/bin/unoconverter &&\
    # unoconverter will look for the Python binary, which has to be at version 3.
    ln -s /usr/bin/python3 /usr/bin/python &&\
    # Verify installations.
    libreoffice --version &&\
    unoconverter --version &&\
    # Cleanup.
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN \
    # Install PDFtk, QPDF & ExifTool (PDF engines).
    # See https://github.com/gotenberg/gotenberg/pull/273.
    curl -o /usr/bin/pdftk-all.jar "https://gitlab.com/api/v4/projects/5024297/packages/generic/pdftk-java/$PDFTK_VERSION/pdftk-all.jar" &&\
    chmod a+x /usr/bin/pdftk-all.jar &&\
    printf '#!/bin/bash\n\nexec java -jar /usr/bin/pdftk-all.jar "$@"' > /usr/bin/pdftk && \
    chmod +x /usr/bin/pdftk &&\
    apt-get update -qq &&\
    apt-get upgrade -yqq &&\
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends qpdf=11.3.0-1+deb12u1 exiftool poppler-utils &&\
    # See https://github.com/nextcloud/docker/issues/380.
    mkdir -p /usr/share/man/man1 &&\
    # Verify installations.
    pdftk --version &&\
    qpdf --version &&\
    exiftool --version &&\
    # Cleanup.
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Support for arbitrary user IDs (OpenShift).
# See:
# https://github.com/gotenberg/gotenberg/issues/1049.
# https://docs.redhat.com/en/documentation/openshift_container_platform/4.15/html/images/creating-images#use-uid_create-images.
RUN \
    usermod -aG root gotenberg &&\
    chgrp -R 0 /home/gotenberg &&\
    chmod -R g=u /home/gotenberg

# Improve fonts subpixel hinting and smoothing.
# Credits:
# https://github.com/arachnys/athenapdf/issues/69.
# https://github.com/arachnys/athenapdf/commit/ba25a8d80a25d08d58865519c4cd8756dc9a336d.
COPY build/fonts.conf /etc/fonts/conf.d/100-gotenberg.conf

# Copy the Golang binaries.
COPY --from=compress-go-binaries-stage /home/pdfcpu /usr/bin/
COPY --from=compress-go-binaries-stage /home/gotenberg /usr/bin/

# Environment variables required by modules or else.
ENV CHROMIUM_BIN_PATH=/usr/bin/chromium
ENV LIBREOFFICE_BIN_PATH=/usr/lib/libreoffice/program/soffice.bin
ENV UNOCONVERTER_BIN_PATH=/usr/bin/unoconverter
ENV PDFTK_BIN_PATH=/usr/bin/pdftk
ENV QPDF_BIN_PATH=/usr/bin/qpdf
ENV EXIFTOOL_BIN_PATH=/usr/bin/exiftool
ENV PDFCPU_BIN_PATH=/usr/bin/pdfcpu
ENV PDFTOPPM_BIN_PATH=/usr/bin/pdftoppm
ENV PDFTOPPM_DPI=203
ENV PDFTOPPM_ANTIALIASING=no

USER gotenberg
WORKDIR /home/gotenberg

# Default API port.
EXPOSE 3000

ENTRYPOINT [ "/usr/bin/tini", "--" ]
CMD [ "gotenberg" ]
