## The Itinerary-Prettifier

### Main

The **main()** function starts out by declaring two boolean variables, both for flags. Then two flag objects are created, the first has the name h and the second has the name d. Both are set by default to false. The job of one is to display the usage of the program the second will display the output of the program on the command line. Then the program parses the input and gives the flags the values that the user wants. If the h flag is razed or the command line has received less than three arguments the usage is displayed and the program stops working.

If all inputs are according to usage the program takes the input, output and airport file paths from the command line, cuts the first ./ and uses them as paths to the files. Next, the program runs the **Validation()** function, if the validation function returns an error the program stops working. If it does not then all the necessary elements are present and the program can function correctly.

Now comes the core of the program itself. The **inputRead()** function is called, it creates the variable *rawText*, containing the raw unchecked slice of strings. Then the **processText()** function is called which creates the *convertedText* variable that contains the checked and altered string. Finally the **outputWrite()** function is called which takes the path given and the checked string and writes it. The final if statement checks if the user wants to also display the output on the command line, checking the *displayFlag* boolean and calling **outputDisplay()** function if true.

### Validation

The **Validation()** function takes two string variables and returns an error. The first variable is the path to the input file the second is the path to the airport variable. At first, the function attempts to read the input file, but since we do not want to use the input string here the function does not read the contents of the file into a variable, but instead, if there is an error it gives the error variable an error value. If the error has a value then the error is returned and the program prints "Input not found." The second block is similar, but the difference is that the function needs to use the data and so it simply opens the file and creates a variable. If there is an error in opening the file the function returns the error and prints "Airport lookup not found."

Moving on the function reads the airport csv and first checks that there are 6 columns, if not the function returns an error, a custom one this time, just for the error to have a value and prints the line "Airport lookup malformed." If the columns are good the function begins a for loop where it checks every cell and if any cell is empty it returns an error and prints "Airport lookup malformed."

### inputRead

The function takes the path to the input file as an argument and returns a slice of strings. It starts with reading into the *rawFile* variable the input, as a byte datatype and then converts the variable into the *stringFile* variable. What follows are three regular expressions and replacements based on those expressions. First, the trailing whitespaces are removed by matching two or more following spaces and replacing them with a single space. Then all the different whitespace characters are converted into newline characters. (**Note:** When writing in a txt file the action of pressing enter produces two whitespace characters \r and \n, I convert the carriage return to a newline because the instructions say that whitespace characters have to be converted and that two new space characters are allowed then every text that has a new line in the file has two new lines in the output file)

Once the formatting is done the large string is split by the newline characters and if the last element of that slice is empty then it is removed. Once done the slice of strings is returned.

### processText

The function takes the raw unchecked slice of strings and the path to the airport csv and returns a string. at first, it declares the slice of strings variable *processed* then it creates a for loop that takes in every element of the slice and further splits it by spaces to create single words. Then a second for loop inside the first is created that iterates over every word in the line. A *punct* variable is created just in case the word has at its end a punctuation mark. Also, a regular expression pattern is created which checks whether the words start with the datetime format. The core of this loop consists of two if checks, if the word does not correspond to either it is not altered. 

The first if statement checks whether the word is in a datetime format and if it is then it checks the last element to see if it is a punctuation mark. If it is then the punctuation mark is saved in the *punct* variable because the conversion process removes the punctuation mark. Then the datetime is fed into the **formatTime()** function. The function returns the correctly formatted time, if there is an error the word is not altered, if there is no error the word is replaced with the reformated one and the punctuation is returned to the end of the word.

The second if statement is similar but without the error statement because if the code is wrong or is not found the function simply returns it. Once the line has been checked it is rebound by spaces and added to the *processed* slice. Once the entire slice of strings is compleated the processed slice will contain all the checked lines in itself. Then it is joined back together by newline characters and returned.

### formatTime

The function takes a single word and returns that word and an error. At first, the *formatted* string variable is declared and that variable will contain the result of the formatting. Then two patterns are created, one for day, the other for the two times. If the word does not match either pattern it is returned as is with an error.

If the word starts with *D* then it is first split by the symbol *T*, making two halves, then the first half is also split based on the dash. Once the two splits are done the day and the year can be extracted from the splits. the month is converted into an integer so that it can be used in the **time.Date()** function, which makes it a time object and which can be formatted into the abbreviations of the months through the **Format()** function. Then all the strings are added together through spaces and returned as a new word.

If the word is in the time format then three variables are declared. The function first checks if the word starts with *T12*, and if it does then gives add the *true* value. Then if the string contains a "+" the word is first split by the plus sign and then the current time and the difference from Zulu time is extracted. If the word contains "-" then the word is split from the minus and the times are extracted. also the neg value is set to *true*. If neither is the case the current time is extracted and the movement is given a correctly formatted 0 value.

Now if the word contains a minus then the direction is set to a minus sign, also the current time is converted into an integer. If add is *true* then an affix variable is declared that contains "AM" and if the hour is larger than 12 then the affix changes to "PM" and the hour is recalculated. Once everything is extracted the new format is put together by adding all the strings together and then returned. Just for the 24-hour format, there are fewer steps. 

### airportRead

The function takes in the path to the file and the word from the loop. It returns the word. First, the unchanged variable is declared that keeps all the values the word has through the process just in case the word is not found in the loop later. The first condition checks if there is a "*" symbol in front of the code and sets the city boolean to *true* if there is. It also removes the symbol from the word. Secondly, it removes the "#" symbols from the word as well. Then the unchanged variable is completed and the csv file is opened and closed. A reader is created and the header variable is created to keep track of the names of the columns. Then also a map is created so. 

Then an infinite *for* loop is created that is broken when the last value in the file is reached. In the loop, there is another loop which loops through the columns and rows of the csv file, creating a map object which has the key value of its header. The *if* statement checks if the values of the row contain the word and are the same length as the word. If it finds a match it replaces the word with the name from the same row. If the city value is *true* the municipality instead of the airport name is used. If the function finds the name it returns the replaced value, if not it returns the unchanged value.

### outputWrite

The function takes both the path and the formated string and does not return anything. First, it creates a file with the path given and then orders the file closed when the function finishes. Then it used **file.WriteString()** function to write into the *.txt* file.


### outputDisplay

This function takes the formatted string. First, it creates a pattern to find the offset time and then matches them in the string, specifying that it should find all matches in the string. Then it creates a *for* loop where it adds characters to the beginning and the end of the matches to make the characters bold. the **QuoteMeta()** function has to be used because the matches contain *regex* sensitive characters. Then it compiles the text together and prints it out into the terminal.



