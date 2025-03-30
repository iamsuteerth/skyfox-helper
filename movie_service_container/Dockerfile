FROM ruby:2.7.4-alpine

RUN apk add --no-cache build-base

WORKDIR /app

COPY Gemfile* ./

RUN bundle install --without development test

COPY . .

RUN rm -rf /usr/local/bundle/cache/*.gem && \
    find /usr/local/bundle/gems/ -name "*.c" -delete && \
    find /usr/local/bundle/gems/ -name "*.o" -delete

EXPOSE 4567

CMD ["bundle", "exec", "rackup", "--host", "0.0.0.0", "-p", "4567"]