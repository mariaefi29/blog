# Personal blog

Source code for a personal blog website with a responsive layout and minimum use of external packages and frameworks.
The application is running now at [http://www.marialife.com](https://www.marialife.com)
## Front end

* HTML
* pure CSS (CSS Grid, flexbox, SVGs)
* vanilla JavaScript

## Backend

* Golang
* MongoDB

## Building an application locally

1. Make sure you have MongoDB database running either locally or remotely. [MongoDB installation process](https://docs.mongodb.com/manual/installation/)
2. Create 3 collections in your db (named "blog"): posts, comments and emails. If you want to populate them (which is not necessary), you can check database schema in the package models.  
3. Create an environmental variable DB_CONNECTION_STRING and assign it to an url of your MongoDB database. For example: mongodb://localhost/blog
4. Create environmental variables SMTP_EMAIL and SMTP_PASSWORD and assign them to an SMTP server credentials. This step is needed for a contact form to work, but is not necessary for the application to run.
5. Run `make` in a command line in a working directory.
6. Open http://localhost:8080/ in your browser and enjoy surfing.

Note: It was my first project in Go. Now when I look at it with some experience, I would like to rewrite almost everything. I am planning to release a major version when I have time to rewrite all the code.
