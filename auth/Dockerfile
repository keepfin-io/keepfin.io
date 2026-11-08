FROM node:26-alpine AS development
WORKDIR /app
COPY package*.json .npmrc ./
RUN npm ci
COPY tsconfig*.json ./
COPY src ./src
COPY test ./test
USER node
CMD ["npm", "run", "dev"]

FROM development AS build
USER root
RUN npm run typecheck && npm run build

FROM node:26-alpine AS production
WORKDIR /app
COPY package*.json .npmrc ./
RUN npm ci --omit=dev
COPY --from=build /app/dist ./dist
USER node
EXPOSE 3001
CMD ["node", "dist/server.js"]
