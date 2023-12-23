FROM gcc as build-backend

RUN apt-get update && apt-get install -y cmake;
RUN apt-get install libpq-dev -y

COPY ./backend /app/src

WORKDIR /app/build

RUN cmake ../src

RUN cmake --build .

FROM node as build-frontend

WORKDIR /app

ADD ./frontend/package.json .
RUN npm i

ADD ./frontend .

RUN npm run build

FROM httpd


RUN mkdir /usr/local/apache2/api

RUN apt-get update -y && apt-get install libpq-dev -y

COPY ./httpd/my-httpd.conf /usr/local/apache2/conf/httpd.conf
COPY --from=build-backend /app/build/lab1 /usr/local/apache2/api/index.cgi
COPY --from=build-frontend /app/dist /usr/local/apache2/htdocs

RUN mkdir /usr/local/apache2/saved_files

COPY ./entrypoint.sh ./

RUN chmod +x entrypoint.sh

CMD ./entrypoint.sh