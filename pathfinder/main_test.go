package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Test that the program can find more than one route for 2 trains between waterloo and st_pancras for the London Network Map (The number of trains can be supstituded on line 37).
func TestLondonNetworkMap(t *testing.T) {

	content := []byte(`stations:
waterloo,3,1
victoria,6,7
euston,11,23
st_pancras,5,15

connections:
waterloo-victoria
waterloo-euston
st_pancras-euston
victoria-st_pancras
`)

	file, err := os.Create("london.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("london.txt")
	cmd := exec.Command("go", "run", "main.go", "london.txt", "waterloo", "st_pancras", "2")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(output), "\n")

	firstLine := strings.Fields(lines[0])
	if len(firstLine) != 2 {
		t.Fatalf("Expected the first line to contain 2 strings but got %d", len(firstLine))
	}
}

// It finds only a single valid route for 1 train between waterloo and st_pancras in the London Network Map.
func TestLondonNetworkMapOne(t *testing.T) {

	content := []byte(`stations:
waterloo,3,1
victoria,6,7
euston,11,23
st_pancras,5,15

connections:
waterloo-victoria
waterloo-euston
st_pancras-euston
victoria-st_pancras
`)

	file, err := os.Create("london.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("london.txt")
	cmd := exec.Command("go", "run", "main.go", "london.txt", "waterloo", "st_pancras", "1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(output), "\n")

	firstLine := strings.Fields(lines[0])
	if len(firstLine) != 1 {
		t.Fatalf("Expected the first line to contain 2 strings but got %d", len(firstLine))
	}
}

// It prints the train movements with the correct format "T1-station", "T2-station" etc.
func TestFormat(t *testing.T) {

	content := []byte(`stations:
waterloo,3,1
victoria,6,7
euston,11,23
st_pancras,5,15

connections:
waterloo-victoria
waterloo-euston
st_pancras-euston
victoria-st_pancras
`)

	file, err := os.Create("london.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("london.txt")
	cmd := exec.Command("go", "run", "main.go", "london.txt", "waterloo", "st_pancras", "1")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal(err)
	}

	words := strings.Split(string(output), " ")

	if words[0][0] != 'T' && words[0][1] != '1' {
		t.Fatalf("Expected the format T!, got %b%b", words[0][0], words[0][1])
	}
}

// Tests that it completes the movements in no more than 6 turns for 4 trains between bond_square and space_port.
func TestAppleOrange(t *testing.T) {
	content := []byte(`stations:
bond_square,20,6
apple_avenue,7,7
orange_junction,6,1
space_port,1,11

connections:
bond_square-apple_avenue
apple_avenue-orange_junction
orange_junction-space_port
`)
	file, err := os.Create("apple.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("apple.txt")
	cmd := exec.Command("go", "run", ".", "apple.txt", "bond_square", "space_port", "4")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal(err)
	}

	var lines []string
	rougtLines := strings.Split(string(output), "\n")
	for _, line := range rougtLines {
		if len(line) >= 1 {
			lines = append(lines, line)
		}
	}

	if len(lines) > 6 {
		t.Fatalf("Expected the total lines to be not larger than 6 but got %d", len(lines))
	}

}

// Tests that it completes the movements in no more than 8 turns for 10 trains between jungle and desert.
func TestJungleDesert(t *testing.T) {
	content := []byte(`stations:
jungle,5,16
green_belt,6,1
village,5,7
mountain,9,16
treetop,0,4
grasslands,15,13
suburbs,4,9
clouds,0,0
wetlands,2,12
farms,11,10
downtown,4,4
metropolis,3,20
industrial,1,18
desert,9,0

connections:
jungle-grasslands
mountain-treetop
clouds-wetlands
downtown-metropolis
green_belt-village
suburbs-clouds
industrial-desert
jungle-farms
village-mountain
wetlands-desert
grasslands-suburbs
jungle-green_belt
farms-downtown
treetop-desert
metropolis-industrial
mountain-wetlands
farms-mountain
`)
	file, err := os.Create("jungle.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("jungle.txt")
	cmd := exec.Command("go", "run", ".", "jungle.txt", "jungle", "desert", "10")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal(err)
	}

	var lines []string
	rougtLines := strings.Split(string(output), "\n")
	for _, line := range rougtLines {
		if len(line) >= 1 {
			lines = append(lines, line)
		}
	}

	if len(lines) > 8 {
		t.Fatalf("Expected the total lines to be not larger than 6 but got %d", len(lines))
	}
}

// Tests that it completes the movements in no more than 11 turns for 20 trains between beginning and terminus.
func TestBeginningTerminus(t *testing.T) {
	content := []byte(`stations:
beginning,0,0
near,1,0
far,1,3
terminus,0,3
connections:
beginning-near
beginning-terminus
near-far
terminus-far
`)
	file, err := os.Create("beginning.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("beginning.txt")
	cmd := exec.Command("go", "run", ".", "beginning.txt", "beginning", "terminus", "20")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal(err)
	}

	var lines []string
	rougtLines := strings.Split(string(output), "\n")
	for _, line := range rougtLines {
		if len(line) >= 1 {
			lines = append(lines, line)
		}
	}

	if len(lines) > 11 {
		t.Fatalf("Expected the total lines to be not larger than 6 but got %d", len(lines))
	}
}

// Tests that it completes the movements in no more than 6 turns for 4 trains between two and four.
func TestTwoFour(t *testing.T) {
	content := []byte(`stations:
one,1,1
two,2,2
three,3,3
four,4,4
five,5,5
six,6,6

connections:
two-three
five-one
three-one
two-five
one-four
six-two
one-six
`)
	file, err := os.Create("two.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("two.txt")
	cmd := exec.Command("go", "run", ".", "two.txt", "two", "four", "4")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal(err)
	}

	var lines []string
	rougtLines := strings.Split(string(output), "\n")
	for _, line := range rougtLines {
		if len(line) >= 1 {
			lines = append(lines, line)
		}
	}

	if len(lines) > 6 {
		t.Fatalf("Expected the total lines to be not larger than 6 but got %d", len(lines))
	}
}

// Tests that it completes the movements in no more than 6 turns for 9 trains between beethoven and part.
func TestBeethovenPart(t *testing.T) {
	content := []byte(`stations:
beethoven,1,6
verdi,7,1
albinoni,1,1
handel,3,14
mozart,14,9
part,10,0

connections:
beethoven-handel
handel-mozart
beethoven-verdi
verdi-part
verdi-albinoni
beethoven-albinoni
albinoni-mozart
mozart-part
`)
	file, err := os.Create("beethoven.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("beethoven.txt")
	cmd := exec.Command("go", "run", ".", "beethoven.txt", "beethoven", "part", "9")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal(err)
	}

	var lines []string
	rougtLines := strings.Split(string(output), "\n")
	for _, line := range rougtLines {
		if len(line) >= 1 {
			lines = append(lines, line)
		}
	}

	if len(lines) > 6 {
		t.Fatalf("Expected the total lines to be not larger than 6 but got %d", len(lines))
	}
}

// Tests that it completes the movements in no more than 8 turns for 9 trains between small and large.
func TestSmallLarge(t *testing.T) {
	content := []byte(`stations:
stations:
small,4,0
large,4,6
00,0,0
01,0,1
02,0,2
03,0,3
04,0,4
05,0,5
10,1,0
11,1,1
12,1,2
13,1,3
14,1,4
15,1,5
20,2,0
21,2,1
22,2,2
23,2,3
24,2,4
25,2,5
30,3,0
31,3,1
32,3,2
33,3,3
34,3,4
35,3,5
36,3,6

connections:
24-25
24-23
23-12
small-32
32-33
33-34
34-35
35-36
36-22
small-10
10-11
10-20
11-12
11-14
12-large
12-03
small-13
13-14
14-15
small-00
00-01
01-02
02-03
03-04
20-21
20-25
21-15
21-22
21-30
22-large
25-30
30-31
31-large
04-05
05-large
`)
	file, err := os.Create("small.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("small.txt")
	cmd := exec.Command("go", "run", ".", "small.txt", "small", "large", "9")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatal(err)
	}

	var lines []string
	rougtLines := strings.Split(string(output), "\n")
	for _, line := range rougtLines {
		if len(line) >= 1 {
			lines = append(lines, line)
		}
	}

	if len(lines) > 8 {
		t.Fatalf("Expected the total lines to be not larger than 6 but got %d", len(lines))
	}
}

// Tests that it displays "Error" on stderr when too few command line arguments are used. It is the same error when end or start stations do not exist, because it would lead to too few arguments.
func TestTooFewArguments(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "waterloo", "st_pancras", "euston", "victoria", "2")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: train scheduler usage:\ngo run . [path to file containing network map] [start station] [end station] [number of trains]\noptional flag -a before other arguments to use distance-based pathfinding\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when too many command line arguments are used.
func TestTooManyArguments(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "london.txt", "waterloo", "st_pancras", "victoria", "2")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: train scheduler usage:\ngo run . [path to file containing network map] [start station] [end station] [number of trains]\noptional flag -a before other arguments to use distance-based pathfinding\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when the start and end station are the same.
func TestSame(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "london.txt", "waterloo", "waterloo", "2")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: start and end station are the same\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when no path exists between the start and end stations.
func TestNoPath(t *testing.T) {
	content := []byte(`stations:
beethoven,1,6
verdi,7,1
albinoni,1,1
handel,3,14
mozart,14,9
part,10,0

connections:
handel-mozart
verdi-part
verdi-albinoni
albinoni-mozart
mozart-part
`)
	file, err := os.Create("beethoven.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("beethoven.txt")
	cmd := exec.Command("go", "run", ".", "beethoven.txt", "beethoven", "part", "9")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: no path found\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when the number of trains is not a valid positive integer.
func TestInteger(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "london.txt", "waterloo", "waterloo", "-2")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: number of trains must be positive int:\n\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when any of the coordinates are not valid positive integers.
func TestCoordinates(t *testing.T) {
	content := []byte(`stations:
waterloo,3,1
victoria,6,-7
euston,11,23
st_pancras,5,15

connections:
waterloo-victoria
waterloo-euston
st_pancras-euston
victoria-st_pancras
`)
	file, err := os.Create("beethoven.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("beethoven.txt")
	cmd := exec.Command("go", "run", ".", "beethoven.txt", "beethoven", "part", "9")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: Invalid Coordinates: victoria\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when two stations exist at the same coordinates.
func TestLocation(t *testing.T) {
	content := []byte(`stations:
waterloo,3,1
victoria,3,1
euston,11,23
st_pancras,5,15

connections:
waterloo-victoria
waterloo-euston
st_pancras-euston
victoria-st_pancras
`)
	file, err := os.Create("beethoven.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("beethoven.txt")
	cmd := exec.Command("go", "run", ".", "beethoven.txt", "beethoven", "part", "9")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: duplicate coordinates: waterloo and victoria\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when a connection is made with a station which does not exist.
func TestConnection(t *testing.T) {
	content := []byte(`stations:
waterloo,3,1
victoria,4,1
euston,11,23
st_pancras,5,15

connections:
waterloo-victoria
waterloo-euston
waterloo-circus
st_pancras-euston
victoria-st_pancras
`)
	file, err := os.Create("beethoven.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("beethoven.txt")
	cmd := exec.Command("go", "run", ".", "beethoven.txt", "beethoven", "part", "9")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: invalid connection: waterloo-circus\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when station names are duplicated.
func TestDuplication(t *testing.T) {
	content := []byte(`stations:
waterloo,3,1
waterloo,3,4
victoria,4,1
euston,11,23
st_pancras,5,15

connections:
waterloo-victoria
waterloo-euston
waterloo-circus
st_pancras-euston
victoria-st_pancras
`)
	file, err := os.Create("beethoven.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("beethoven.txt")
	cmd := exec.Command("go", "run", ".", "beethoven.txt", "beethoven", "part", "9")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: duplicate station name--waterloo\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when station names are invalid.
func TestInvalidNames(t *testing.T) {
	content := []byte(`stations:
waterloo,3,1
victoria,4,1
euston,11,23
st_pancras,5,15

connections:
waterloo-victoria
waterloo-euston
st_pancras-euston
victoria-st_pancras
`)
	file, err := os.Create("beethoven.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("beethoven.txt")
	cmd := exec.Command("go", "run", ".", "beethoven.txt", "beethove", "part", "9")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: Start or end station not found.\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when the map does not contain a "stations:" section.
func TestStations(t *testing.T) {
	content := []byte(`connections:
waterloo-victoria
waterloo-euston
st_pancras-euston
victoria-st_pancras
`)
	file, err := os.Create("beethoven.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("beethoven.txt")
	cmd := exec.Command("go", "run", ".", "beethoven.txt", "beethove", "part", "9")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: stations section missing\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when the map does not contain a "connections:" section.
func TestConnections(t *testing.T) {
	content := []byte(`stations:
waterloo,3,1
victoria,4,1
euston,11,23
st_pancras,5,15
`)
	file, err := os.Create("beethoven.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("beethoven.txt")
	cmd := exec.Command("go", "run", ".", "beethoven.txt", "beethove", "part", "9")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}

		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: connections section missing\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}

// Tests that it displays "Error" on stderr when a map contains more than 10000 stations.
func TestTooMany(t *testing.T) {
	var content strings.Builder
	content.WriteString("stations:\n")
	for i := 0; i < 10001; i++ {
		content.WriteString(fmt.Sprintf("station%d,%d,%d\n", i, i, i))
	}
	content.WriteString("connections:\n")

	file, err := os.Create("toomany.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.WriteString(content.String()); err != nil {
		t.Fatal(err)
	}

	file.Close()
	defer os.Remove("toomany.txt")
	cmd := exec.Command("go", "run", "main.go", "toomany.txt", "start", "end", "2")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		// Check for the expected error message in stderr
		errMsg := strings.TrimSpace(stderr.String())
		expectedErrMsg := "Error: more than 10,000 stations--10001\nexit status 1"
		if errMsg != expectedErrMsg {
			t.Errorf("Expected error message: '%s', got: %s", expectedErrMsg, errMsg)
		}
	} else {
		t.Fatal("Expected an error but got none.")
	}
}
