# MongoDB

Containerised mongoDB single node instance

## Configuration

[Create security keyfile](https://docs.mongodb.com/manual/tutorial/deploy-replica-set-with-keyfile-access-control/#std-label-deploy-repl-set-with-auth) for the replicas to authenticate eachother with
```
openssl rand -base64 756 > mykey.key
```

```
chmod 400 mykey.key
```

```
sudo chown 999:999 mykey.key
```

Connect with initialsed root user details
```
mongosh -u root -p rootpassword
```
Initiate the replica set
```
rs.initiate()
```
[Change config so replica set members work correctly](https://stackoverflow.com/a/41067404)
```
cfg = rs.conf()
```

```
cfg.members[0].host = "localhost:27017"
```

```
rs.reconfig(cfg)
```


Add a new user for creating users and roles
```
use admin
db.createUser(
  {
    user: "userAdmin",
    pwd: passwordPrompt(),
    roles: [ { role: "userAdminAnyDatabase", db: "admin" } ]
  }
)
```

Disconnect and then login to the admin user
```
use admin
db.auth("userAdmin", passwordPrompt())
```

Create a new user to read and write to `testdb`
```
use testdb
db.createUser(
  {
    user: "workerUser",
    pwd: passwordPrompt(),
    roles: [ { role: "readWrite", db: "testdb" }],
  }
)
```

Insert data into collection `dayta` in `testdb`
```
db.dayta.insertOne({name:"Cathal"})
```

Change root user password
```
mongosh admin -u root -p rootpassword
db.changeUserPassword("root", passwordPrompt())
```

Create a user for the metrics exporter
```
use admin
db.createUser(
  {
    user: "mongodb_exporter",
    pwd: passwordPrompt(),
    roles: [
        { role: "clusterMonitor", db: "admin" },
        { role: "read", db: "local" }
    ]
  }
)
```
Place the mongoDB exporter user password into an env
```
echo "MONGODB_CONNECT_URL=mongodb://mongodb_exporter:[PASSWORDHERE]@mongodb:27017" > .env
```
Restart both containers to let the mongoDB exporter connect

### Creating Collections With Schemas

First marshal the struct into JSON. Make *sure* all json keys are all lowercase before proceeding to create the schema.

Use [JSON Schema](https://www.jsonschema.net/home) to translate the JSON to a valid mongo DB schema. However, 

* All `$id`, `example` and `default` tags must be removed. 
* All `integer` tags must also be replaced with `number`


```
db.createCollection("collectionName", {validator: {$jsonSchema:{....}}})
```
