  
FROM ubuntu:20.04
ENV TZ=America/New_York
ENV PATH=/go/bin:$PATH

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone &&\
    apt update &&\
    apt -y install gdal-bin gdal-data libgdal-dev

RUN mkdir -p /app
COPY ./go-fema-consequences /app/
WORKDIR /app/
RUN mkdir /app/media

ENTRYPOINT ["/app/go-fema-consequences"]