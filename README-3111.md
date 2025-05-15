# Debugging Issue 3111

- Run the proxy on the host: `go run helpers/proxy/main.go`
  - Turn off the firewall / open the port!
- Modify apt.debughttp and replace the HOST with the IP of the Docker host
- `docker build -t debian:gopass -f Dockerfile.debian .`
- `docker run --rm -ti debian:gopass`
- Inside the container:
- `apt update`
