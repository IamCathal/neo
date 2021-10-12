# Workflows Documentation

## Get a new node up and running

To get a new node up and running there are a few key things that must be setup first.

- Create a new user with minimal privelages
- Docker and github actions runner must be installed and configured with appropriate tag to reference in the workflow
- Github actions runner must be running


### Create a new user
```
adduser ghrunner
```
```
usermod -aG sudo ghrunner
```
```
sudo su - ghrunner
```


### Install docker and github actions runner
```
curl -fsSL https://get.docker.com -o get-docker.sh
```
```
sh get-docker.sh
```
```
docker -v
```
```
sudo apt update
```
```
sudo apt install docker-compose
```
```
docker-compose -v
```
```
logout
```
```
sudo groupadd docker
```
```
sudo usermod -aG docker ghrunner
```

Now setup the github runner [official instructions here](https://github.com/IamCathal/neo/settings/actions/runners/new)
