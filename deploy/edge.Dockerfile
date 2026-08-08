FROM node:22-alpine AS frontend-build

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM node:22-alpine AS admin-build

WORKDIR /app/admin
ARG VITE_BASE=/admin/
ENV VITE_BASE=$VITE_BASE
COPY admin/package*.json ./
RUN npm ci
COPY admin/ ./
RUN npm run build

FROM nginx:1.27-alpine

COPY deploy/nginx.prod.conf /etc/nginx/conf.d/default.conf
COPY --from=frontend-build /app/frontend/dist /usr/share/nginx/html
COPY --from=admin-build /app/admin/dist /usr/share/nginx/html/admin

EXPOSE 80
