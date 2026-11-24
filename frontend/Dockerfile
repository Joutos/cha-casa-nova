# syntax=docker/dockerfile:1
# ---------- build ----------
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY . .
ARG VITE_PIX_KEY=""
ARG VITE_PIX_QR_IMAGE=""
ENV VITE_PIX_KEY=$VITE_PIX_KEY VITE_PIX_QR_IMAGE=$VITE_PIX_QR_IMAGE
RUN npm run build

# ---------- runtime: Nginx sem root, servindo a SPA e fazendo proxy da API ----------
FROM nginxinc/nginx-unprivileged:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
