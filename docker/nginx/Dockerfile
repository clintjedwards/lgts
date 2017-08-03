FROM nginx:stable-alpine
MAINTAINER Clint Edwards <clint.j.edwards@gmail.com>

RUN rm /etc/nginx/conf.d/default.conf
COPY ./docker/nginx/nginx.conf /etc/nginx/conf.d/lgts.conf

COPY ./docker/nginx/lgts.crt /etc/ssl/certs/lgts.crt
COPY ./docker/nginx/lgts.key /etc/ssl/private/lgts.key

ENTRYPOINT ["nginx", "-g", "daemon off;"]