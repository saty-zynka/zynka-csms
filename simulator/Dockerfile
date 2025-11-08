# Multi-stage build for SolidStudio VCP Simulator
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Production stage
FROM node:18-alpine

WORKDIR /app

# Copy node_modules and source from builder
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./
COPY --from=builder /app/tsconfig.json ./
COPY --from=builder /app/src ./src
COPY --from=builder /app/index_16.ts ./
COPY --from=builder /app/index_201.ts ./
COPY --from=builder /app/index_21.ts ./
COPY --from=builder /app/index_16_2_connectors.ts ./

# Install tsx globally for running TypeScript
RUN npm install -g tsx

# Environment variables (can be overridden)
ENV WS_URL=ws://gateway:9311/ws/cs001
ENV CP_ID=cs001
ENV PASSWORD=
ENV ADMIN_PORT=9999
ENV OCPP_VERSION=16

# Expose admin port
EXPOSE 9999

# Default to OCPP 1.6, but allow override via OCPP_VERSION env var
ENTRYPOINT ["sh", "-c"]
CMD ["if [ \"$OCPP_VERSION\" = \"201\" ]; then npx tsx index_201.ts; elif [ \"$OCPP_VERSION\" = \"21\" ]; then npx tsx index_21.ts; else npx tsx index_16.ts; fi"]

