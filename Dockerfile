FROM scratch
EXPOSE 9470
WORKDIR /
COPY cachet_exporter .
ENTRYPOINT ["./cachet_exporter"]