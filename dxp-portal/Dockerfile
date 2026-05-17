FROM node:20-bookworm-slim

ENV NODE_ENV=production

WORKDIR /app

RUN apt-get update && apt-get install -y python3 make g++ && rm -rf /var/lib/apt/lists/*

COPY packages/backend/dist/bundle.tar.gz app-config*.yaml ./
RUN tar xzf bundle.tar.gz && rm bundle.tar.gz

COPY node_modules/ ./node_modules/

RUN cd /app/node_modules/isolated-vm && npm rebuild

CMD ["node", "packages/backend", "--config", "app-config.yaml"]
