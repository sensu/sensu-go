# proto2xsd

XSD conversion tool for Google ProtocolBuffers version 3

	> proto2xsd -help
		Usage of proto2xsd [flags] [path ...]
  		-ns string
    		namespace of the target types (default "http://your.company.com/domain/version")		
  		-w	write result to XSD files instead of stdout

## Docker
A Docker image is available on [DockerhuB](https://hub.docker.com/r/emicklei/proto2xsd/).
It can be used as part of your continuous integration build pipeline.

### build 
	GOOS=linux go build
	docker build -t emicklei/proto2xsd:latest -t emicklei/proto2xsd:0.2 .

### run
	 docker run -v $(pwd):/data emicklei/proto2xsd -ns http://company/your/v1 /data/YOUR.proto

© 2017, [ernestmicklei.com](http://ernestmicklei.com).  MIT License.     