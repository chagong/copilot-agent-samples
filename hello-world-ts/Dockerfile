FROM node:lts-alpine
RUN mkdir -p /home/node/app/node_modulesp
WORKDIR /home/node/app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE 3000
CMD [ "node", "dist/index.js" ]