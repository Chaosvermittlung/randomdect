package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var errormsg string
var execdir string

type frontpage struct { //Datentyp der an den template parser übergeben wird
	ErrormsgHTML template.HTML //Fehlermeldung als HTML string
	Extension    int
	Name         string
	Called       int
}

func recieveHandler(w http.ResponseWriter, r *http.Request) {

	file, header, err := r.FormFile("file") // the FormFile function takes in the POST input id file

	if err != nil { //Fehler beim erstellen der datei
		//fmt.Fprintln(w, err)
		errstring := err.Error()
		errormsg = "<div class=\"alert alert-danger\" role=\"alert\"><font color=\"#FF0000\"><b>Error!</b> Unable to form file: " + errstring + "</font></div>"
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	defer file.Close()

	fmt.Println(filepath.Ext(header.Filename))

	err = os.MkdirAll(execdir+"/uploaded/", 0777)
	if err != nil {
		errstring := err.Error()
		errormsg = "<div class=\"alert alert-danger\" role=\"alert\"><font color=\"#FF0000\"><b>Error!</b> Unable to create file: " + errstring + "</font></div>"
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	out, err := os.Create(execdir + "/uploaded/" + header.Filename) //Speicher file in uploaded
	if err != nil {                                                 //Fehler beim abspeichern
		errstring := err.Error()
		errormsg = "<div class=\"alert alert-danger\" role=\"alert\"><font color=\"#FF0000\"><b>Error!</b> Unable to create file: " + errstring + "</font></div>"
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		errstring := err.Error()
		errormsg = "<div class=\"alert alert-danger\" role=\"alert\"><font color=\"#FF0000\"><b>Error!</b> " + errstring + "</font></div>"
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	fmt.Println("Info:", "File uploaded successfully:", header.Filename)

	p, err := loadPhonebook(execdir + "/uploaded/" + header.Filename)
	if err != nil {
		errstring := err.Error()
		errormsg = "<div class=\"alert alert-danger\" role=\"alert\"><font color=\"#FF0000\"><b>Error!</b> " + errstring + "</font></div>"
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	err = p.Insert()

	if err != nil {
		errstring := err.Error()
		errormsg = "<div class=\"alert alert-danger\" role=\"alert\"><font color=\"#FF0000\"><b>Error!</b> " + errstring + "</font></div>"
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	errormsg = "<div class=\"alert alert-success\" role=\"alert\">File uploaded</div>"
	http.Redirect(w, r, "/", http.StatusFound)
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	exts := r.FormValue("extension")

	ext, err := strconv.Atoi(exts)
	if err != nil {
		errstring := err.Error()
		errormsg = "<div class=\"alert alert-danger\" role=\"alert\"><font color=\"#FF0000\"><b>Error!</b> " + errstring + "</font></div>"
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	fmt.Println(action, exts)
	switch action {
	case "called":
		err = increasecalled(ext)
	case "optout":
		err = optout(ext)
	case "delete":
		err = remove(ext)
	default:
		err = errors.New("Wrong action:" + action)
	}

	if err != nil {
		errstring := err.Error()
		errormsg = "<div class=\"alert alert-danger\" role=\"alert\"><font color=\"#FF0000\"><b>Error!</b> " + errstring + "</font></div>"
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	errormsg = "<div class=\"alert alert-success\" role=\"alert\">Action completed</div>"
	http.Redirect(w, r, "/", http.StatusFound)
}

func mainHandler(w http.ResponseWriter, r *http.Request) { //Bedient http request auf /
	myfrontpage := frontpage{}

	ee, err := getEntries(1)
	if err != nil {
		fmt.Println("There was an error:", err)
		return
	}
	sel := rand.Intn(len(ee))
	if ee[sel].Called > 0 {
		reroll := ee[sel].Called
		origreroll := reroll
		for (reroll > 0) && (ee[sel].Called >= origreroll) {
			sel = rand.Intn(len(ee))
			reroll = reroll - 1
		}
	}

	myfrontpage.Extension = ee[sel].Extension
	myfrontpage.Name = ee[sel].Name
	myfrontpage.Called = ee[sel].Called

	t, err := template.ParseFiles(execdir + "/templates/main.html") //Lade Template
	if err != nil {
		fmt.Println("There was an error:", err)
		return
	}

	myfrontpage.ErrormsgHTML = template.HTML(errormsg)

	err = t.Execute(w, &myfrontpage) //Zeige template an
	if err != nil {
		fmt.Println("There was an error:", err)
		return
	}
	errormsg = ""
}

func main() {
	initialisation()
	rand.Seed(time.Now().UTC().UnixNano())
	var err error
	execdir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	//Demo Server Handler
	//Setzt alle Routen zu den entsprechenden Handlern
	r.HandleFunc("/", mainHandler)
	r.HandleFunc("/recieve", recieveHandler)
	r.HandleFunc("/set", setHandler)

	log.Fatal(http.ListenAndServe(":4242", r)) //auf port 4242 lauschen
}
