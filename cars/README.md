# The Cars API, Server, and Webpages

## Setup

Clone the repository without version control metadata:

```text
git clone https://gitea.kood.tech/markokarilaid/cars && rm -rf /.git
```

Extract the API files into `./api` 
```text
unzip api.zip
```

Ensure the file structure looks like so:
```
cars/
│
├── api/
│   ├── Makefile/
│   ├── main.js/
|   └──...
├── main.go
├── README.md
└── ...
```

Make sure you have installed and updated [NodeJS](https://nodejs.org/en) and [NPM](https://www.npmjs.com/package/npm).
Run the servers from the `/cars` directory:

```text
go run main.go
```
The main function will initialize the API server at localhost:3000 and our server at localhost:4444. There may be errors thrown, they are (probably) ignorable if you at the end you see

`Starting server on localhost:4444`

You are now ready to check out the website!

# The Site


Use a web browser to navigate to 
`http://localhost:4444`. 

## Welcome Page
On first visit you will have three options:
- Free-text search for a car by name, manufacturer, or category
- View all cars
- View the list of manufacturers 

On further visits, and after engaging the website in certain ways, you will see an additional header with your top 3 most viewed cars.

## Car List Table
Using the search or view all will take you to `http://localhost:4444/list` where you will see a table of cars matching your criteria. From here you can interact in the following ways:
- Click the "Return to Home" button
- Click on the car name to see a popup with detailed specifications
- Click the manufacturer name to see more details
- Select two cars for comparison using the checkbox in the "Compare" column, then click the "Compare (select two)" button 

## Compare Page
Once you've selected the two cars to compare and hit the compare button, you are taken to `http://localhost:4444/list/compare` where you will see a side-by-side detailed comparison of various features and specifications. At the bottom of the page are buttons to return to the Home page or the List. 

## Recommendations 
After some experimentation, return to the home page to see your "recommended cars" i.e. your most viewed cars. The following events trigger a view count increment:
- Viewing car details by clicking the car name on the list page
- Viewing manufacturer details by clicking the manufacturer on the list page
- Comparing the car to another

The recommendations are currently configured to only last one session. 
