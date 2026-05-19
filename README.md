# How to run migrations on database
If you haven't installed mongoose, run:
```
go install github.com/pressly/goose/v3/cmd/goose@latest

// Optional, to export the path in case you can't use goose
export PATH=$PATH:$HOME/go/bin
source ~/.zshrc
```
To check if you have mongoose installed, run ```goose -version```.

To install postgresql, run:
```
brew install postgresql
brew services start postgresql@18
```
To check if you have postgresql installed, run ```psql --version```.

To create a new migrations, run:
```
cd migrations
goose create <migration description> sql
```

Then, run:
```
cd ../
export DATABASE_URL=<your database uri>
goose -dir migrations postgres "$DATABASE_URL" up 
```

# Run Docker
To build the image for the application:
```
docker build --tag bettero-app .
docker images # To see the list of images
```

To create the container fo the image:
```
docker run \               
    -d \
    -p 8080:8080 \
    -e "DATABASE_URL=<database_uri>" \
    --name <container's name> \
    bettero-app

docker stop <containder's name> # Stop the container
docker start <container's name> # Start the container
```