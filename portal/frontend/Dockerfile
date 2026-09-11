FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
RUN chmod -R 644 /usr/share/nginx/html && chmod -R +X /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
