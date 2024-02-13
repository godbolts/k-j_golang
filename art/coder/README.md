## Art-Decoder

### Main

The **main()** function starts out by declaring three variables, *encodedText* to store the input, *multiLine* as a multiline input flag and *encoder* as an encoding flag. Then it defines the boolean values as flags and sets their default as *false.* The first if statement checks if the *multiLine* flag has been raised and based on that the input is read in either as a single line argument through the **flag.Arg()** of a multiline input through the **input()** function. The second if statement checks if the *encoding* flag has been raised. If it has then the **reEncoder()** function is used and the result is printed out. If the flag has not been raised the **validity()** function is used to check the input, if the function returns an error the program stops working. Then the *slicer()** function is used to return a slice of strings, where some are the chunks of encoded information. finally the **capture()** function is used to convert the encoded blocks into art. The last line prints out the image.

### Input

The **input()** function uses the bufio.NewReader function to read from user input. It stores every line as a slice in the *slicedText* variable. It stops reading once it encounters two newline characters. Then it joins all the slices together if they are not empty and removes newline characters if they are the last characters in the string.

### ReEncoder

The **reEncoder()** is a function that has two parts. The first loop checks if there are repeated single symbols and the second loop checks if there are repeated pairs of symbols. The first loop goes through the input string and uses the cache variable to temporarily store and compare strings. If the cache is empty or if the symbol in there is the same the looping symbol is added to the cache. But if the looping symbol is not the same as the one in the cache and the cache is not empty then it checks the size of the cache. If the cache is larger than one it encodes the cache with the **numberSymbol()** function adds it to the sliced array, clears the cache and adds the new symbol to the cache. If the cache has the length of one then the same thing happens but without the encoding function. Last is an if check on whether the loop is on its last element and if it is it checks the length of the cache and stores the value appropriately.

The second loop is similar but instead of single symbols, it checks pairs. The only meaningful difference is the beginning if statements that check whether the second element of the pairs is out of index bounds and if it is the function acts appropriately. Since the loop functions by adding to j by 2, the function has to function differently if the string is made out of even or non-even numbers. Another difference is that the **doubleSymbol()** function is used in the loop because the length of the symbol is double the size and the occurrence of it has to be half the size of the length.

### Validity

The **validity()** function just checks if the input has an equal amount of opening and closing brackets in itself. it does so by storing the number of brackets in two variables *open* and *closed.* Then it loops through the input and stores the amount of brackets it encountered in the variables and lastly, it checks whether they have the same amount. If not it produces an error.

### Capture

The capture function takes in a slice and creates a pattern that can capture elements that start and close with square brackets. Then it loops through the slice. If there is a match the loop gets the number of occurrences and the length of the number from the brackets with the **getNumbers()** function. If the space between the number and the symbol is not a space it returns an error. The symbol is captured through indexes, using the length of the number and excluding the last element of the string. If there is no symbol in the brackets and the loop picks up a bracket the loop returns an error. Then a new element is made that uses the **strings.Repeat()** function which then contains the n number of the symbol in it. Finally, the whole slice is joined back together with the decode strings added to the unaffected ones.

### GetNumbers

The function takes in the encoded string. In the loop, it skips the first element because it is a bracket and adds the second element to the variable *number* if the next element is not a digit it breaks the loop. This way the function can get as large a number as it needs and only breaks when the number ends. Once the number has been stored its length is stored in the *lenght* variable that is necessary for properly slicing the string later and the number itself is converted into an integer. Both numbers are returned. Here it also checks whether the encoding has a number in it because if it does not the function cannot convert it and returns an error.