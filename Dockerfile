FROM scratch
ADD jobs-server /
ADD cfg/jobs-mapping.json cfg/
CMD ["/jobs-server"]
EXPOSE 8080 
