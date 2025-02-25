# https://github.com/citusdata/pg_cron/issues/17#issuecomment-2405879858
FROM postgres:16-alpine

RUN \
  apk update && \
  apk upgrade && \
  apk add --no-cache postgresql-pg_cron=1.6.2-r0 --repository=https://dl-cdn.alpinelinux.org/alpine/v3.20/community;

RUN ln -s /usr/lib/postgresql16/pg_cron.so /usr/local/lib/postgresql/pg_cron.so && \
  ln -s /usr/share/postgresql16/extension/pg_cron* /usr/local/share/postgresql/extension

CMD [\
  "postgres",\
  "-c",\
  "shared_preload_libraries=pg_cron"\
]