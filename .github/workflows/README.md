# Workflows Documentation

## Get a new node up and running

To get a new node up and running there are a few key things that must be setup first.

- Create a new user with minimal privelages
- Docker and github actions runner must be installed
- 


#### Create a new user
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


#### Install docker and github actions runner
Get docker:
```
# Curl the install script
curl -fsSL https://get.docker.com -o get-docker.sh
# Install docker
sh get-docker.sh
# Try it out
docker -v
```
Get docker-compose
```
# Update package manager
sudo apt update
# Get new docker-compose
sudo apt install docker-compose
# Try it out
docker-compose -v
```
[Official instructions here](https://github.com/IamCathal/neo/settings/actions/runners/new)