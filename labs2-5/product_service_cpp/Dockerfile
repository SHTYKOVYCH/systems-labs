FROM gcc as build

RUN apt-get update && apt-get install -y cmake;

ADD . /app/src

WORKDIR /app/build

RUN cmake ../src

RUN cmake --build .

FROM httpd

COPY --from=build /app/build/lab1 /usr/local/apache2/cgi-bin/

CMD httpd-foreground -c "LoadModule cgid_module modules/mod_cgid.so"
