FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy worker script
COPY worker.js .

# Run worker
CMD ["node", "worker.js"]