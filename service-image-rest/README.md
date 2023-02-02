# Image Service
The image service is a REST API service which storage and load image file.

## Testing
Then run command:

```shell
curl -vv localhost:9090/1/go.mod -X PUT --data-binary @test.png
```