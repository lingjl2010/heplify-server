FROM csighub.tencentyun.com/tccc/golang:1.24.2-bookworm-tccc
ENV GOPROXY=https://goproxy.woa.com,direct GOSUMDB="sum.woa.com+643d7a06+Ac5f5VOC4N8NUXdmhbm8pZSXIWfhek5JSmWdWrq7pLX4"
RUN sed -r 's/(deb|security).debian.org/mirrors.tencent.com/' -i /etc/apt/sources.list.d/debian.sources && \
    apt-get update && apt-get install -y libavcodec-dev && apt-get clean
