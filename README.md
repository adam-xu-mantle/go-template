# Mantle Kratos Project Template

## Install Kratos
```
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
```
## Create a service
```
# Create a template project
# create project's layout
kratos new {{project-name}} -r https://github.com/adam-xu-mantle/go-template

cd {{project-name}}

```
## Compilation and Running
```
make init
make all
kratos run -w . 
```
## Try it out
```
curl 'http://127.0.0.1:8000/helloworld/mantle'

The response should be
{
  "message": "Hello mantle"
}
wire
```

## Docker
```bash
# build
docker build -t <your-docker-image-name> .

# run
docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/your/configs>:/data/conf <your-docker-image-name>
```

