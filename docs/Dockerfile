FROM klakegg/hugo:ext-alpine

RUN apk add git npm

COPY package.json package-lock.json ./

RUN npm install
