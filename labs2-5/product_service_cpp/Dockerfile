FROM gcc as build

RUN apt-get update && apt-get install -y cmake;

ADD . /app/src

WORKDIR /app/build

RUN cmake ../src

RUN cmake --build .

FROM httpd


RUN mkdir /usr/local/apache2/api
COPY ./httpd/my-httpd.conf /usr/local/apache2/conf/httpd.conf
COPY --from=build /app/build/lab1 /usr/local/apache2/api/index.cgi

