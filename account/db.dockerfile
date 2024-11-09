FROM postgres:13.1

COPY ./account/up.sql /docker-entrypoint-initdb.d/up.sql

CMD ["postgres"]