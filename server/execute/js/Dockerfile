FROM node:14

WORKDIR /app

COPY  ./execute/js/code.js .

COPY ./execute/js/package-code.json ./package.json

RUN npm install

CMD ["node", "code.js"]
