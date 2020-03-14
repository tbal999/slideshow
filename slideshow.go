package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type slideshow struct {
	Title      string
	Header     string
	Subheader  []string
	Subbheader string
	Code       string
	Image      string
}

func Check(filename string) int {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	r := csv.NewReader(f)
	r.Comma = ' '
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	_, err2 := r.Read()
	if err2 != nil {
		return 1
	}
	return 0
}

func openbrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

//Ensures that a DIR exists
func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModeDir)
	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

func WriteToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}

func (s slideshow) presento(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./ui/slideshow.gtpl")
	HomePageVars := s
	for index := range HomePageVars.Subheader {
		HomePageVars.Subbheader += "- " + HomePageVars.Subheader[index] + "\n" + "\n"
	}
	if err != nil { // if there is an error
		log.Print("template executing error: ", err) //log it
	}
	err = t.Execute(w, HomePageVars)
	if err != nil { // if there is an error
		log.Print("template executing error: ", err) //log it
	}
}

func start(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("method:", r.Method) //get request method
	listy := []string{}
	file, err := os.Open("slideshow.txt")
	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		listy = append(listy, scanner.Text())
	}
	var x = 0
Stage1:
	S := slideshow{} //store the details
	for x < len(listy) {
		s := strings.Split(listy[x], "_")
		switch s[0] {
		case "S":
			S.Title = strings.Join(s[1:], " ")
			x++
		case "**":
			S.Header = strings.Join(s[1:], " ")
			x++
		case "***":
			S.Subheader = append(S.Subheader, strings.Join(s[1:], ""))
			x++
		case "-":
			S.presento(w, r)
			x++
			goto Stage1
		case "IMG":
			S.Image = strings.Join(s[1:], "")
			x++
		case "TXT":
			tobeopened := strings.Join(s[1:], "")
			file2, err := os.Open(tobeopened)
			if err != nil {
				fmt.Println(err)
			}
			scanner := bufio.NewScanner(file2)
			for scanner.Scan() {
				S.Code += scanner.Text() + "\n"
			}
			x++
		case "E":
			S.Title = "Q & A"
			S.presento(w, r)
			x++
		}
	}
}

func main() {
	ensureDir("ui")
	filerrorA := WriteToFile("./ui/slideshow.gtpl", `<title>Slideshow</title>
</head>
<body>
<style>
body {
  background-color: white;
}

h1 {
  color: black;
}
h2 {
  color: black;
  margin-left: 75px;
  margin-right: 75px;
}
h3 {
  color: black;
  margin-left: 150px;
}
.borderexample {
    background-color: #dfe8e3;
    border-radius: 5px;
}
p {
    white-space: pre-wrap;
}
</style>
  <br>
  <br>
<h1> <center><b><font size="+5">{{.Title}}</size></b></center></font></h1> 
<h2 class="borderexample"><center><b><font size="+3">{{.Header}}</size></b></center>
<font size="+2">
<p>
<center><img src="/public/{{.Image}}" onerror="this.style.display='none'"/></img></center>
<p pre-wrap>{{.Subbheader}}</p>
<p pre-wrap>{{.Code}}</p>
</h2>
<center>
<font color="white">
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p>
<p>.</p></center>
</body>`)
	if filerrorA != nil {
		log.Fatal(filerrorA)
	}
	http.HandleFunc("/", start)
	fmt.Println("Loading default browser...")
	fmt.Println("Just press CTRL-C to exit when you have finished")
	openbrowser("http://127.0.0.1:9090")
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./"))))
	err2 := http.ListenAndServe(":9090", nil) // setting listening port
	if err2 != nil {
		log.Fatal("ListenAndServe: ", err2)
	}
}
