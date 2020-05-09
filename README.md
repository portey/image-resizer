# Image resizer

#### Run service
```
make dockerise 
make docker-up
``` 

#### In order to upload a new image, use this curl: 
```
curl http://localhost:8130/query \
  -F operations='{"query":"mutation ($file: Upload!) { uploadImage(image:$file, sizes:[{ width:100, height:100 }]) { id  path  clientName  mimeType  size  uploadAt  sizes {    path    width    height  }  } }", "variables": { "file": null } }' \
  -F map='{ "0": ["variables.file"] }' \
  -F 0=@./resizer/fixtures/image.jpg
```