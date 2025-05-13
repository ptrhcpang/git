package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func read_file_to_bytestring(sha string) []byte{

	path := fmt.Sprintf(".git/objects/%v/%v", sha[:2], sha[2:])
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %s\n", err)
	}
	r, _ := zlib.NewReader(io.Reader(file))
	s, _ := io.ReadAll(r)
	r.Close()

	return s
}


func sha1_encoder(blobtext []byte) []byte{

	sha := sha1.New()
	sha.Write(blobtext)
	encrypted := sha.Sum(nil)

	return encrypted

}


func write_blob(filename string) []byte{
	filetext, _ := os.ReadFile(filename)
	filetextLength := []byte(strconv.Itoa(len(filetext)))
	byteSlice := [][]byte{[]byte("blob "), filetextLength, []byte("\x00"), filetext}
	separator := []byte("")
	blobtext := bytes.Join(byteSlice, separator)

	return blobtext
}


func make_gitobject(encrypted []byte, blobtext []byte){

	//use sha key to make new directory and file
	encryptedString := fmt.Sprintf("%x",encrypted)
	
	dir := ".git/objects/" + encryptedString[:2]
	objName := dir + "/" + encryptedString[2:]

	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
	}

	//zlib compress text and write to 
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(blobtext)
	w.Close()

	if err := os.WriteFile(objName, b.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
	}		

}


func read_tree(s []byte, nameonlyFlag bool) {

	sLength := len(s)
	nameBuffer := make([]byte, 260)
	nameLen := 0

	if (!bytes.Equal(s[:4],[]byte("tree"))){
		//error
		fmt.Fprintf(os.Stderr, "Error: not a tree object.")
	}

	current_index := 4

	for (s[current_index] != 0x0){
		current_index += 1
		//current index now at first \x00
	}

	readFlag := 0 //(mod 3) 0-reading mode, 1-reading name, 2-reading sha

	for (current_index < sLength){

		current_index += 1

		switch readFlag%3{
		case 0://mode
			if(bytes.Equal(s[current_index:current_index + 6],[]byte("40000 "))){
				if(!nameonlyFlag){fmt.Print("040000 tree ")}
				current_index -= 1
			}else if(bytes.Equal(s[current_index:current_index + 6],[]byte("100644")) && !nameonlyFlag){
				fmt.Print("100644 blob ")
			}else if(bytes.Equal(s[current_index:current_index + 6],[]byte("100755")) && !nameonlyFlag){
				fmt.Print("100755 exec ")
			}else if(bytes.Equal(s[current_index:current_index + 6],[]byte("120000")) && !nameonlyFlag){
				fmt.Print("120000 link ")
			}else{
				if(!nameonlyFlag){
					fmt.Print("xxxxxx err- ")			
				}
			}
			current_index += 7
			readFlag += 1
		
		case 1://name
			nameLen = 0;
			current_index -= 1
			for (s[current_index]!= 0x0){
				nameBuffer[nameLen] = s[current_index]
				nameLen += 1
				current_index +=1
			}
			readFlag += 1
		
		case 2://sha key and print
			if(!nameonlyFlag){
				fmt.Print(hex.EncodeToString(s[current_index:current_index+20]))
				fmt.Print("    ")
			}
			fmt.Print(string(nameBuffer[:nameLen]),"\n")
			current_index += 19
			readFlag += 1

		}
	}
}


func write_tree() []byte{

	dir_list, err := os.ReadDir(".")
	var buffer []byte
	var sc_tree []byte
	var encrypted []byte
	var byteSlice [][]byte
	separator := []byte("")

	if err != nil {
        panic(err)
    }

	for _, entry := range(dir_list){

		if(strings.Compare(entry.Name(),".git")==0){
			
			continue

		}else if(entry.IsDir()){

			//begin the tree entry
			byteSlice = [][]byte{buffer, []byte("40000 "), []byte(entry.Name()), []byte("\x00")}
			buffer = bytes.Join(byteSlice, separator)

			//write byte slice of entire tree file of sub-directory
			os.Chdir(entry.Name())
			sc_tree = write_tree()
			
			//encode the byte slice 
			encrypted = sha1_encoder(sc_tree)

			//append the sha1 code to the first part of the treee entry
			byteSlice = [][]byte{buffer, encrypted}
			buffer = bytes.Join(byteSlice, separator)
			os.Chdir("..")

		}else{

			//read file and write blob plaintext
			filetext := write_blob(entry.Name())
			
			//calculate file sha
			encrypted = sha1_encoder(filetext)
			byteSlice = [][]byte{buffer, []byte("100644 "), []byte(entry.Name()), []byte("\x00"), encrypted}
			buffer = bytes.Join(byteSlice, separator)

		}

	}
	
	//append file header to the head of file text
	treelength := []byte(strconv.Itoa(len(buffer)))
	byteSlice = [][]byte{[]byte("tree "), treelength, []byte("\x00"), buffer}
	buffer = bytes.Join(byteSlice, separator)
	
	//return entire file
	return buffer
}


func write_commit(tree_hash []byte, parent_hash []byte, message []byte) []byte{

	separator := []byte("")
	temp := time.Now()
	now := []byte(strconv.FormatInt(temp.Unix(),10))

	//UTC offset as byte string
	var sign string
	_, here_sec := temp.Zone()
	if(here_sec >= 0){
		sign = " +"
	}else{
		sign = " -"
	}

	if(here_sec < 0){
		here_sec = - here_sec
	}
	if(here_sec < 36000){
		sign = sign + "0"
	}
	here := []byte(sign + string(here_sec/3600 + 48) + "00")

	var a_seconds, a_timezone []byte 

	//author, committer information
	author := []byte("author Josephine March <jo.march@math.mit.edu> ")
	a_seconds = now
	a_timezone = here

	committer := []byte("committer Susan Pevensie <susan.pevensie@st-annes.ox.ac.uk> ")
	c_seconds := now
	c_timezone := here

	
	//file header header
	//separator := []byte("")
	byteSlice := [][]byte{[]byte("tree "), tree_hash, []byte("\n") }
	buffer := bytes.Join(byteSlice, separator)

	//parents
	byteSlice = [][]byte{buffer, parent_hash}
	buffer = bytes.Join(byteSlice, separator)

	//insert author information
	byteSlice = [][]byte{buffer, author, a_seconds, a_timezone, []byte("\n")}
	buffer = bytes.Join(byteSlice, separator)

	//insert committer information
	byteSlice = [][]byte{buffer, committer, c_seconds, c_timezone, []byte("\n\n")}
	buffer = bytes.Join(byteSlice, separator)

	//insert message
	byteSlice = [][]byte{buffer, message, []byte("\n")}
	buffer = bytes.Join(byteSlice, separator)

	//append file header to the head of file text
	commitlength := []byte(strconv.Itoa(len(buffer)))
	byteSlice = [][]byte{[]byte("commit "), commitlength, []byte("\x00"), buffer}
	buffer = bytes.Join(byteSlice, separator)
	
	//return entire file text
	return buffer

}


// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	
	fmt.Fprintf(os.Stderr, "Logs from your program will appear here!\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":

		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}
		
		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}
		
		fmt.Println("Initialized git directory")
	
	
	case "cat-file":

		sha := os.Args[3]

		s := read_file_to_bytestring(sha)
		parts := strings.Split(string(s), "\x00")

		if(os.Args[2] == "-t"){
			fmt.Print(parts[0][:4])
		}else{
			fmt.Print(parts[1])
		}

	case "hash-object":

		filename := os.Args[3]

		plain_blob := write_blob(filename)
		encrypted := sha1_encoder(plain_blob)
		make_gitobject(encrypted, plain_blob)

		encryptedString := fmt.Sprintf("%x",encrypted)
		fmt.Println(encryptedString)

		
	case "ls-tree":

		tree_sha := os.Args[2]
		nameonlyFlag := false

		if(os.Args[2] == "--name-only"){
			nameonlyFlag = true
			tree_sha = os.Args[3]
		}

		s := read_file_to_bytestring(tree_sha)
		read_tree(s,nameonlyFlag)
	

	case "write-tree":

		plain_tree := write_tree()
		encrypted := sha1_encoder(plain_tree)
		make_gitobject(encrypted, plain_tree)
	
		encryptedString := fmt.Sprintf("%x",encrypted)
		fmt.Println(encryptedString)	

	
	case "commit-tree":
		
		//converts hex string to byte slice
		_, err1 := hex.DecodeString(os.Args[2])
		tree_hash := []byte(os.Args[2])
		if(err1 != nil || len(os.Args[2]) != 40){
			panic("Tree hash is incorrect.")
		}

		//parse remainder of inputs
		var message, parent []byte
		parent_hash := []byte{}
		separator := []byte("")
		var byteSlice [][]byte
		var err2 error
		i := 0

		//parse parent/commit hashes and message
		for (i < len(os.Args)){

			if(os.Args[i] == "-m" && i + 2 == len(os.Args)){
				message = []byte(os.Args[i + 1])
				i = i + 1
			}else if(os.Args[i] == "-p" && i + 1 < len(os.Args)){
				_, err2 = hex.DecodeString(os.Args[i + 1])
				parent = []byte(os.Args[i + 1])
				if(err2 != nil || len(os.Args[i + 1]) != 40){
					panic("Commit hash is incorrect.")
				}else{
					byteSlice = [][]byte{parent_hash, []byte("parent "), parent, []byte("\n")}
					parent_hash = bytes.Join(byteSlice,separator)
				}
				i = i + 1
			}
			
			i = i + 1
		}

		//write commit object to file
		plain_commit := write_commit(tree_hash, parent_hash, message)
		encrypted := sha1_encoder(plain_commit)
		make_gitobject(encrypted, plain_commit)
	
		encryptedString := fmt.Sprintf("%x",encrypted)
		fmt.Println(encryptedString)

	case "clone":

		

	default:
		
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

