FROM debian:bookworm-slim
SHELL ["/bin/bash", "-c"]
RUN apt update && apt upgrade -y
RUN apt install curl unzip -y

# install node 20
RUN curl -sL https://deb.nodesource.com/setup_20.x | bash -
RUN apt update && apt upgrade -y
RUN apt install nodejs -y

# install bun
RUN curl -fsSL https://bun.sh/install | bash
ENV PATH "$PATH:/root/.bun/bin"

# install the app
ADD . /app
WORKDIR /app
RUN bun install --no-save --frozen-lockfile
RUN bun run --filter @metal/webapp build
CMD bun run --filter @metal/webapp start
