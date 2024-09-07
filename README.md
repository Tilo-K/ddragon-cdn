
<p align="center">

<img src="https://raw.githubusercontent.com/Tilo-K/ddragon-cdn/main/logos/logo.png" style="width: 300px"/>

</p>

<div style="display: flex; gap: 1rem;">

[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)
![GitHub Issues or Pull Requests](https://img.shields.io/github/issues/Tilo-K/ddragon-cdn)

</div>
This project can be used to host the latest ddragon assets provided by Riot Games. The server checks for new versions and updates the assets accordingly.


## Deployment

To deploy this project locally run:

```bash
  go build -o cdn
  ./cdn
```

To deploy this project via docker:

```bash
docker pull ghcr.io/tilo-k/ddragon-cdn:main
docker run -it -p 3000:3000 -e PORT=3000 -e STORAGE_DIR=/data -v .\data:/data ghcr.io/tilo-k/ddragon-cdn:main
```
## Environment Variables

To run this project, you will need to add the following environment variables to your docker container / system

`STORAGE_DIR` to specify where the files should be stored.

`PORT` to specify the port on which the server listens. Default: 6002
## Contributing

Contributions are always welcome!

Just create a PR if you have somethinh usefull to add.
Alternativly create an issue if there is something you would like to see.

Please adhere to this project's [code of conduct](https://tilok.dev/coc).


## License

[MIT](https://choosealicense.com/licenses/mit/)

