# img-proxy

img-proxy is a proof of concept project designed to create a scalable, distributed image cache proxy system.
It utilizes hashing to distribute images across a cluster and the
[HashiCorp memberlist library](https://github.com/hashicorp/memberlist) for managing cluster membership.

## How it works
A client sends an HTTP GET request to the only public accessible endpoint which looks like the following:
`gatewaynode-N:8080/image?url="..."` (replace gatewaynode-N with your gateway instance host)
Since there may be multiple gateway nodes, it's advisable to place an HTTP load balancer in front of them.
Alternatively, you can opt for a single instance deployment. If you choose to use a load balancer, ensure that the
gateway nodes are not directly accessible except through the load balancer.

## Gateway Node
A gateway node performs a straightforward operation: it computes the hash of an image and subsequently dispatches the
task to the corresponding worker node. Currently, this process relies on the calculation of the modulo of the image hash,
ensuring that each image is assigned to a specific worker node. To offer more flexibility and less redistributing when
the worker nodes count changes, consistent hashing should be implemented instead.

## Worker Node
Worker nodes are exclusively accessed by gateway nodes and should not be accessible from any others. When a gateway
node sends a request to a worker node, it first checks whether the requested image is already stored in the local worker
node cache. If the image is not found, it is downloaded from a third-party source and then cached locally. To further
enhance performance, additional optimizations such as image compression and resizing could be implemented. Additionally,
it is advisable to set memory limits or define cache invalidation times for the system.

## Why & Use case
I always wanted to get my hands on a distributed, scalable and containerized project and after reading the
[Discord Blog post](https://discord.com/blog/how-discord-resizes-150-million-images-every-day-with-go-and-c) about how
they handle image scaling at mass, I got inspired to create a similar thing.
I want to clarify that **this project is not intended for production use**; it's primarily a proof of concept aimed at
exploring distributed systems. The Discord Blog post did a great job on explaining their use case for such a project.
But I will give you some more use cases, I can think of:
 - you need to provide deterministic and fast access to third party images
 - you want to reduce traffic on your network by resizing, compressing and caching images close to your backend
 - you want to create thumbnails or preview images for the real image. Like a URL preview in a chat app

## Running a dev cluster
Since my goal is to keep it simple, the whole project can be built and run with only two commands.

For building the whole project: `docker compose build`

and running: `docker compose up`

By default, it starts with only one gateway and worker node. To start the cluster with multiple nodes just use the awesome
[docker compose scale](https://docs.docker.com/reference/cli/docker/compose/up/) feature:

`docker compose up --scale gateway=3 --scale worker=3`

Nodes are configured with environment variables, so take a look at the docker-compose.yml file.

## Results

I use [Prometheus](https://prometheus.io/) to collect metrics from our software, demonstrating the disadvantages of
using SHA-256 for data distribution. Despite being a cryptographically secure algorithm, SHA-256 does not ensure uniform
distribution, making it unsuitable for equal data distribution. Additionally, any change in the cluster count
invalidates
all data, highlighting its inflexibility. Consistent hashing offers a better solution to handle variable-sized clusters
more effectively.
![visualizing distribution among clusters](https://raw.githubusercontent.com/phips4/img-proxy/main/docker/grafana%20dashboard.png)

## Endpoints overview
| direction         | request                    | response                                         | description                                         |
|-------------------|----------------------------|--------------------------------------------------|-----------------------------------------------------|
| user -> gateway   | GET /image?url=...         | OK (image) or Bad Request, Internal Server Error | endpoint for users                                  |
| gateway -> worker | GET /v1/image?url          | OK (image) or Not Found                          | if not cached return not found, return cached image | 
| gateway -> worker | POST /v1/cache {"url":...} | OK (image) or Bad Request, Internal Server Error | download and cache image (resize, compression)      | 

