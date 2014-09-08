# docker-tools

A set of tools to make using Docker a bit nicer.


## dbuild

`dbuild` helps when building Docker containers.  Given a root path, a path to a Dockerfile, and an
output file, it will copy the Dockerfile to the root directory, run the build, and then export the
final container to a tar file (optionally compressed).  The Dockerfile is removed from the root
after building (unless it's already located there).

Example:

```
$ dbuild --name myapp build/subdir/Dockerfile build/ myapp.image
2014/01/01 13:14:15 Started
[...]
2014/01/01 13:14:43 Completed successfully

# Copy the image to the Docker host, and then...
$ docker load < myapp.image
```


## dcontrol

`dcontrol` helps when controlling multiple Docker containers.  Given a configuration file in YAML
format, it will resolve dependencies amongst containers and start each container in the correct
order, properly linking containers to each other.  It allows for starting, stopping, and restarting
containers in proper dependency order.


### Configuration Format

TODO
