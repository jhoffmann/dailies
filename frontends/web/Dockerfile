# Build stage
FROM node:current-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY src ./src
COPY public ./public
COPY *.json ./

# Build the Angular app for production
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built app from build stage
COPY --from=builder /app/dist/dailies-angular/browser /app

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
