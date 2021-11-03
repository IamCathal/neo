# MongoDB

Containerised mongoDB single node instance

## Configuration

Connect with initialsed root user details
```
mongosh -u user -p password
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
mongosh admin -u root -p password
db.changeUserPassword("root", passwordPrompt())
```

### Creating Collections With Schemas

Use [JSON Schema](https://www.jsonschema.net/home) to translate JSON to a valid mongo DB schema. However, all `$id`, `example` and `default` tags must be removed. All `integer` tags must also be replaced with `number`

```
db.createCollection("collectionName", {validator: {$jsonSchema:{....}}})
```
