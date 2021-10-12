# Workflows Documentation

## Get a new node up and running

To get a new node up and running there are a few key things that must be setup first.

- Create a new user with minimal privelages
- Docker and github actions runner must be installed
- 


### Create a new user
Add new user
```
adduser ghrunner
```
Add user to sudo group
```
usermod -aG sudo ghrunner
```
Sign in as ghrunner
```
sudo su - ghrunner
```


### Install docker and github actions runner
Get docker:
```
curl -fsSL https://get.docker.com -o get-docker.sh
```
```
sh get-docker.sh
```
```
docker -v
```
Get docker-compose
```
sudo apt update
```
```
sudo apt install docker-compose
```
```
docker-compose -v
```
Sign back into root instead of ghrunner
```
logout
```
```
sudo groupadd docker
```
Add ghrunner to docker group
```
sudo usermod -aG docker ghrunner
```

Now setup the github runner [official instructions here](https://github.com/IamCathal/neo/settings/actions/runners/new)
