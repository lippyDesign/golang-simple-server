FROM golang:1.9-alpine3.7
LABEL maintainer ="V. Lipunov <volo_lipu@yahoo.com>"
# create app directory and change into it
WORKDIR /app
# create env variable for directory of the docker image
ENV SOURCES /go/src/github.com/lippyDesign/golang-simple-server/
# copy everything from current directory to the docker image directory
COPY . ${SOURCES}
# set env port
ENV PORT 8080
# image will expose port 8080
EXPOSE 8080
# change directory into the docker image and build
RUN cd ${SOURCES}; go build -o myapp; cp myapp /app/
# command to run to start the program
ENTRYPOINT ["./myapp"]