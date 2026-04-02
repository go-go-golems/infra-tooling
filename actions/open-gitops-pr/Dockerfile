FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        git \
        gh \
        python3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /action

COPY entrypoint.sh /action/entrypoint.sh
COPY src /action/src

RUN chmod +x /action/entrypoint.sh

ENTRYPOINT ["/action/entrypoint.sh"]
