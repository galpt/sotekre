FROM node:18-alpine AS base
WORKDIR /app
COPY package*.json ./
RUN npm ci --no-audit --no-fund
COPY . .
EXPOSE 3000
CMD ["npm", "run", "dev"]
