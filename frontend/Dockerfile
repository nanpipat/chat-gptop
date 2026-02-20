FROM node:20

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json* ./

RUN npm install

COPY frontend/ .

RUN npm run build

CMD ["npm", "start"]
