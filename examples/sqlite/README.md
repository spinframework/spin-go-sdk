# Requirements
- [**go**](https://go.dev/dl/) - v1.25+
- [**spin**](https://github.com/spinframework/spin) - Latest version
- [**componentize-go**](https://github.com/bytecodealliance/componentize-go) - Latest version

# Usage

The first time you run the example you will need to create the database. Spin will do this for you when passing the `--sqlite` flag and referencing the sql file.

```console
spin build --up --sqlite @db/pets.sql
```

After the database is created you can run Spin as usual:

```console
spin build --up
```

In another terminal, you'll interact with the Spin app:

```console
curl localhost:3000
```
