FROM node:20-alpine

WORKDIR /app

# Install build deps for grpc binaries used by fabric-network
RUN apk add --no-cache python3 make g++ git

COPY package.json ./
RUN npm install --omit=dev

COPY src ./src

ENV NODE_ENV=production
EXPOSE 3000

CMD ["node", "src/server.js"]
