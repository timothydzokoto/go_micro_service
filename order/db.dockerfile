FROM postgres:13.1

COPY ./order/up.sql /docker-entrypoint-initdb.d/up.sql

CMD ["postgres"]