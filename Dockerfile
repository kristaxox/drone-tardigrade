# build stage
FROM golang:alpine AS build-env
RUN apk --no-cache add git
ADD . /src
RUN cd /src && go build -o dt ./...

# final stage
FROM alpine
COPY --from=build-env /src/dt /bin/dt
RUN apk -Uuv add ca-certificates
ENTRYPOINT /bin/dt