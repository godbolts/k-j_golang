## Kood/Jõhvi Golang Module

In Kood/jõhvi, the first module teaches the Golang programming language. The aim is to teach the students how to construct backend programs and integrate them with simple front-end solutions. The module is divided into six tasks. The first three are individual and the last three are group tasks. The module starts with creating programs that manipulate strings and ends up with a fully functional forum website.

### First Task **the Itinerary-Prettifier**

It is a command line tool, which reads a text-based itinerary from a file (input), processes the text to make it customer-friendly, and writes the result to a new file (output). The tool converts airport codes into airport names and also converts dates and times that are in in ISO 8601 standard to customer-friendly dates and times.

### Second Task **the Art-Decoder**

It is a command line tool which converts art data into text-based art. The tool both encodes and decodes art. The decoder converts symbols in square brackets into repetitions of those symbols the structure is [number symbol], for example [5 #] turns into #####. It can also encode repeated symbols to make the image smaller.

### Third Task **the Art-Interface**

It is a web interface for the *art-decoder*. It is a server based on Golang which makes it possible to use a web interface to input text and display the resulting artwork. The server communicates with the command line of the system that the server is running on.

### Fourth Task **the Cars-Viewer**

It is a website that showcases information about different car models, their specifications and their manufacturers. The server draws from an API that is stored in the project as well. The API is small and the website is more of a proof of concept because of the small amount of data involved. 

### Fifth Task **the Stations-Pathfinder**

It is a pathfinding program which uses two algorithms to determine the most efficient paths on a map and then to dispatch the trains on the determined paths without crashing them. There are four inputs, the railway map file, the first station, the last station and the number of trains. 

### Sixth Task **the Literary-Lions-Forum**

It is a digital forum that allows for user registration, post creation, sharing and reaction. It is a program built on a SQLite database. It consists of three layers, the first is a function that creates a database file, the second layer is an API that pulls and pushes data from and into that database and a main server that serves between the front-end HTML and the back-end API. 
