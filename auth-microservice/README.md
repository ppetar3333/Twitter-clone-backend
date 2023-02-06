#### Inside docker-compose.yml file change volumes
> /Users/ad/Desktop/authdata/db:/data/db ---> your_path_where_to_store_data_from_database


#### Download and install packages and dependencies
> go get

#### Build tag twitter-server
> docker build --tag twitter-server .

#### Build containers
> docker-compose up --build 

> optional - docker-compose up --build container_name

#### Start containers
> docker-compose up

#### Remove containers
> docker-compose down

#### List images
> docker ps

#### Delete data from auth database
> delete 'authdata' folder from desktop
