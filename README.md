# go-docker-mysql

This project supply a docker setup with a mysql db using docker-compose, and a csv parser in golang which import in base the data extracted from those *.csv.  
The goal is to import timestamp/value couples in a mysql database.

## Docker
First, go in *projetQOS > docker-qos* and run `docker-compose build`, then `docker-compose up`.<br>
It will set up a local docker with a mysql base installed: "QOSenergy".<br/>
This one has 2 users set up, "root" and "user" with a default password "qos".

To access the visual mysql manager, navigate to `http://localhost:8080/`.<br/>
There is no need to set up a specific table for the future import, since the golang script check if the correct table exist and create it if not.

## Parser go
Now, go in *go-app* and run `go run parser.go datas/`.<br/>
`parser.go` is the script to parse CSV files then insert the extracted data in the previous database.<br/>
`datas/` is a folder with some example to parse. 

## References
[docker image](https://hub.docker.com/_/mysql/)<br/>
[docker-compose](https://docs.docker.com/compose/)<br/>
[package mysql for golang](https://godoc.org/github.com/go-sql-driver/mysql)<br/>