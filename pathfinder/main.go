package main

import (
	"bufio"
	"container/heap"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Represents a station in our map.
type Station struct {
	Name     string
	Distance int
	Train    Train
	Occupied bool
	X        int
	Y        int
}

// Train struct used for scheduling.
type Train struct {
	Number int
	Path   int
	Turn   int
}

// Railway map containing both all stations and the connections between them.
type RailMap struct {
	Stations    []*Station
	Connections map[*Station][]*Station
}

// Item struct used in the shortest path algorithm.
type Item struct {
	value    *Station
	priority int
	index    int
}

// Creates the priority queue item.
type PriorityQueue []*Item

// Function that builds all the stations and the railway map.
func buildStations(filePath string) ([]Station, RailMap) {
	var stations []Station
	connections := RailMap{
		Stations:    make([]*Station, 0),
		Connections: make(map[*Station][]*Station),
	}
	mapFile, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return stations, connections // Return an empty slice or handle the error
	}
	defer mapFile.Close()
	scanner := bufio.NewScanner(mapFile)
	stationSection, connectionsSection := false, false
	stationSectionExists := false
	for scanner.Scan() {
		//build stations until "connections:" is hit
		line := scanner.Text()
		if line == "stations:" {
			stationSectionExists = true
			stationSection = true
			continue
		}
		if !stationSectionExists {
			errCall("Error: stations section missing")
		}
		if line == "connections:" {
			//switch to building connections
			stationSection, connectionsSection = false, true
			continue
		}
		line = trimLine(line)
		if line == "" {
			continue
		}
		if stationSection {
			station := makeStation(line)
			checkDuplicates(station, stations)
			stations = append(stations, station)
			connections.Stations = append(connections.Stations, &stations[len(stations)-1])
		} else if connectionsSection {
			connections = addConnection(line, stations, connections)
		}
	}
	// check stations: and connections: sections exist, then for <=10,000 stations
	if !(!stationSection && connectionsSection) {
		errCall("Error: connections section missing")
	} else if len(stations) > 10000 {
		errCall("Error: more than 10,000 stations--" + fmt.Sprint(len(stations)))
	}

	return stations, connections
}

// Used in the buildStations function to construct individual stations.
func makeStation(line string) Station {
	parts := strings.Split(line, ",")
	name := parts[0]
	x, err1 := strconv.Atoi(parts[1])
	y, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || x < 0 || y < 0 {
		errCall("Error: Invalid Coordinates: " + name)
	}
	if name == "" {
		errCall("Error: Invalid name")
	}
	station := Station{
		Name:     parts[0],
		X:        x,
		Y:        y,
		Distance: 1 << 20,
		Occupied: false,
	}
	return station
}

// Used in the buildStations function to add connections to the railway map.
func addConnection(line string, stations []Station, connections RailMap) RailMap {
	stops := strings.Split(line, "-")
	stop1 := stationLookup(stops[0], stations)
	stop2 := stationLookup(stops[1], stations)
	if stop1 == nil || stop2 == nil {
		errCall("Error: invalid connection: " + line)
	}
	// check here for redundant or reverse connections
	checkDupConnections(stop1, stop2, connections)

	connections.Connections[stop1] = append(connections.Connections[stop1], stop2)
	connections.Connections[stop2] = append(connections.Connections[stop2], stop1)
	return connections
}

// Creates a pointer for stations.
func stationLookup(name string, stations []Station) *Station {
	for i := range stations {
		if stations[i].Name == name {
			return &stations[i] // Return the address of the found station
		}
	}
	return nil
}

// Processes raw lines from network map.
func trimLine(line string) string {
	parts := strings.Split(line, "#")
	line = strings.ReplaceAll(parts[0], " ", "")
	return line
}

// Erroro checking function that checks for duplicate names and coordinates and returns an error if it finds one.
func checkDuplicates(station Station, stations []Station) {
	for _, check := range stations {
		if check.Name == station.Name {
			errCall("Error: duplicate station name--" + check.Name)
		} else if check.X == station.X && check.Y == station.Y {
			errCall("Error: duplicate coordinates: " + check.Name + " and " + station.Name)
		}
	}
}

// Checks for duplicated/reversed connections and quits with error if found.
func checkDupConnections(stop1 *Station, stop2 *Station, connections RailMap) {
	for _, check := range connections.Connections[stop1] {
		if check == stop2 {
			errCall("Error: duplicate connection--" + stop1.Name + " and " + stop2.Name)
		}
	}
	for _, check := range connections.Connections[stop2] {
		if check == stop1 {
			errCall("Error: duplicate connection--" + stop1.Name + " and " + stop2.Name)
		}
	}
}

// Prints error string to os.Stderr and calls os.Exit(1).
func errCall(err string) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// These add functions to the PriorityQueue item, think of them like methods in Python classes, functions inside a specific datastructure.
func (pq PriorityQueue) Len() int           { return len(pq) }                         // Returns the lenght of the item.
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].priority < pq[j].priority } // Compares the values of items in the queue.
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }            // Swaps the values in the queue.

// Pushes items into the queue.
func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Item)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

// Pops items out of the queue.
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // Mark as removed
	*pq = old[0 : n-1]
	return item
}

// A dijkstra algorithm that has a n A* booleon, the review cases work with the normal dijkstra because they do not incorporate distance.
func findShortestPath(start *Station, end *Station, connections RailMap, aStar bool, short bool, single []Station) []Station {
	count := 0
	openSet := make(PriorityQueue, 0) // The set of travelable stations.
	heap.Init(&openSet)
	heap.Push(&openSet, &Item{
		value:    start,
		priority: 0,
		index:    count,
	})

	cameFrom := make(map[*Station]*Station) // The path that has already been taken, stores stations.
	gScore := make(map[string]int)          // How much distance does it take to get to this station from start.
	fScore := make(map[*Station]int)        // The sum of the distance from start and to the end station.

	for _, station := range connections.Stations { // Making the stations distance near infinite value before they are evaluated.
		gScore[station.Name] = 1 << 20
		fScore[station] = 1 << 20
	}
	gScore[start.Name] = 0
	fScore[start] = h(start.X, start.Y, end.X, end.Y)

	for openSet.Len() > 0 {
		current := heap.Pop(&openSet).(*Item).value

		//if end, return. also checks that not finding same path of length=1
		if current == end && !(short && cameFrom[current].Name == start.Name) {
			path := reconstructPath(cameFrom, end)
			// when MULTI searching, check that this isn't the shortest path
			if len(path) == len(single) {
				clearStations(connections)
				continue
			}

			return path
		} else if cameFrom[current] != nil {
			if current == end && short && cameFrom[current].Name == start.Name {
				// when we hit the end directly from the start a second time, skip to avoid redundancy
				continue
			}
		}

		for _, neighbor := range connections.Connections[current] {
			tempGScore := gScore[current.Name] + 1
			if (tempGScore < gScore[neighbor.Name] && !neighbor.Occupied) || (short && neighbor.Name == end.Name) || (single != nil && neighbor.Name == end.Name) {
				cameFrom[neighbor] = current

				gScore[neighbor.Name] = tempGScore
				fScore[neighbor] = tempGScore + h(neighbor.X, neighbor.Y, end.X, end.Y)

				// Check if neighbor is not already in the openSet
				found := false
				for _, item := range openSet {
					if item.value == neighbor {
						found = true
						break
					}
				}

				if !found {
					count++
					// check for a* flag to switch to distance heuristic. default is fixed-block
					if aStar {
						heap.Push(&openSet, &Item{
							value:    neighbor,
							priority: fScore[neighbor],
							index:    count,
						})
					} else {
						heap.Push(&openSet, &Item{
							value:    neighbor,
							priority: gScore[neighbor.Name],
							index:    count,
						})
					}
				}
			}
		}
	}

	return nil // No path found
}

// Once a path has been found the history of stations is reconstructed from the cameFrom variable and a functional path is creted.
func reconstructPath(cameFrom map[*Station]*Station, current *Station) []Station {
	path := make([]Station, 0)
	path = append(path, *current)
	current = cameFrom[current]

	for current != nil {
		path = append(path, *current)
		current.Occupied = true
		current = cameFrom[current]
	}
	// Reverse the path before returning
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// Function for calculating Manhattan distance for the A* algorithm.
func h(x1, y1, x2, y2 int) int {
	return abs(x1-x2) + abs(y1-y2)
}

// Absolute value, there can be no negative numbers in the summs.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Function identifies all the possible paths that there can be and assigns a path to every train. It also tracks how many unique paths there are, which informs how many trains can be realeased in one turn.
func findPaths(start *Station, end *Station, connections RailMap, aStar bool, numTrains int, single []Station) ([][]Station, int) {

	var paths [][]Station
	var uniquePaths int
	var short bool
	counter := 0
	netTrains := numTrains

	path := findShortestPath(start, end, connections, aStar, false, single)
	if path == nil {
		if single != nil {
			return nil, 1
		}
		errCall("Error: no path found")
	}
	numTrains--
	paths = append(paths, path)
	if len(path) == 2 {
		short = true // Marks that we have a path directly from start to end for later pathfinding.
	}
	for {
		path := findShortestPath(start, end, connections, aStar, short, single)
		// Dispatch efficiency logic:
		if len(path)-len(paths[0]) < numTrains {
			if len(path) != 0 {
				paths = append(paths, path)
				uniquePaths = len(paths)
			} else if len(paths[counter])-len(paths[0]) < numTrains {
				paths = append(paths, paths[counter]) // Once all the new paths are found all the other trains will be assigned the already existing paths, from most efficient to least.
				counter++
				if counter == uniquePaths {
					counter = 0
				}
			} else {
				paths = append(paths, paths[0])
				counter = 0
			}
			numTrains--
			if numTrains <= 0 {
				break
			}
		} else {
			break
		}
	}
	if uniquePaths == 0 {
		// Check for more optimal multi-pathing options.
		clearStations(connections)
		multiPaths, uniquePaths := findPaths(start, end, connections, aStar, netTrains, paths[0])
		if uniquePaths == 1 {
			return paths, uniquePaths
		}
		if runSchedule(multiPaths, uniquePaths, true) < (len(paths[0]) + netTrains - 1) {
			return multiPaths, uniquePaths
		}
		uniquePaths = 1
	}
	return paths, uniquePaths // Returns a path for every train and the amount of paths that can be started per turn.
}

// If multible paths are more efficient then a single most efficient path then this function clears the connection map of occupancy for the recount.
func clearStations(railMap RailMap) {
	for _, station := range railMap.Stations {
		station.Occupied = false
	}
	for _, stations := range railMap.Connections {
		for _, station := range stations {
			station.Occupied = false
		}
	}
}

// For the updateActiveStations it retrieves the index value of a specific station.
func findStationIndex(stations []Station, targetName string) int {
	for i, station := range stations {
		if station.Name == targetName {
			return i
		}
	}
	return -1 // Not found
}

// This function switches the name of a trains station to the name of its neighbor, whilst checking that no two stations are occupied at the same time.
func updateActiveStations(currentStation []string, path []Station, active [][]string) []string {

	if currentStation == nil {
		currentStation = []string{path[0].Name}
	}
	index := findStationIndex(path, currentStation[len(currentStation)-1])

	if index+1 < len(path) {
		currentStation = append(currentStation, path[index+1].Name)
		//check for identical simultaneous paths
		same := false
		for _, path := range active {
			if len(path) != len(currentStation) {
				continue
			}
			same = true
			for i, stop := range path {
				if currentStation[i] != stop {
					same = false
					break
				}
			}
		}
		if same {
			return nil
		}
		return currentStation
	} else {
		return []string{"*"} // Current station is the last station
	}
}

// Creates an array of arrays that are trains at the first station, then loops through that array until every train is in the end station and prints the state of the network at every turn.
func runSchedule(paths [][]Station, uniquePaths int, counting bool) int {
	track := uniquePaths
	active := make([][]string, len(paths))
	var done bool
	turnCount := 0

	for turn := 0; ; turn++ {
		done = true
		if turn != 0 {
			track = track + uniquePaths
			if track > len(paths) {
				track = len(paths)
			}
		}
		for i := 0; i < track; i++ {
			if len(active[i]) > 0 {
				if active[i][0] == "*" {
					continue
				}
			}
			active[i] = updateActiveStations(active[i], paths[i], active)
			done = false
		}
		if done {
			break
		}

		for i := 0; i < len(active); i++ {
			if len(active[i]) != 0 {
				if active[i][0] != "*" {
					if !counting {
						fmt.Printf("T%d-%s ", i+1, active[i][len(active[i])-1])
					}
				}
			}
		}
		if !counting {
			fmt.Println("")
		} else {
			turnCount++
		}
	}
	return turnCount
}

func main() {
	// Flag for A* distance-based heuristic.
	var aStar bool
	flag.BoolVar(&aStar, "a", false, "use A*")
	flag.Parse()

	// Assess arguments.
	args := os.Args
	if !((len(args) == 5 && !aStar) || (len(args) == 6 && aStar)) {
		errCall("Error: train scheduler usage:\ngo run . [path to file containing network map] [start station] [end station] [number of trains]\noptional flag -a before other arguments to use distance-based pathfinding")
	}
	argShift := 0
	if len(args) == 6 {
		argShift = 1
	}
	networkMap, startName, endName, trainsToRun := args[1+argShift], args[2+argShift], args[3+argShift], args[4+argShift]
	numTrains, err := strconv.Atoi(trainsToRun)
	if numTrains <= 0 || err != nil {
		errCall("Error: number of trains must be positive int:\n")
	}
	if startName == endName {
		errCall("Error: start and end station are the same")
	}

	// Build your slice of stations and the map.
	stations, connections := buildStations(networkMap)

	start, end := stationLookup(startName, stations), stationLookup(endName, stations)
	if start == nil || end == nil {
		errCall("Error: Start or end station not found.")
	}

	// Find efficient paths and dispatch trains.
	paths, uniquePaths := findPaths(start, end, connections, aStar, numTrains, nil)

	// Run and print the line by line run.
	runSchedule(paths, uniquePaths, false)
}
