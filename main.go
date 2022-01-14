package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	yaml "gopkg.in/yaml.v2"
)

var doc = flag.String("doc", "rlogs", "The name of the document to be produced")
var solo = flag.Bool("solo", false, "Blank forms, rider only")
var class = flag.String("class", "", "The class numbers to be selected. Default=all")
var blanks = flag.Int("blanks", 0, "Print <n> blanks only")
var entrant = flag.String("entrant", "", "The entrant numbers to be selected. Default=all")
var outputfile = flag.String("to", "output.html", "Output filename")

var DBH *sql.DB
var OUTF *os.File

var CFG struct {
	EventDate string `yaml:"eventDate"`
	Database  string `yaml:"database"`
}

type Entrant struct {
	EntrantID    int
	Bike         string
	BikeReg      string
	RiderName    string
	RiderFirst   string
	RiderIBA     string
	PillionName  string
	PillionFirst string
	PillionIBA   string
	OdoKms       int
	Class        int
	Phone        string
	Email        string
	NokName      string
	NokRelation  string
	NokPhone     string
	RiderLast    string
	HasPillion   bool
	IsBlank      bool
	EventDate    string
}

func newEntrant() *Entrant {

	var e Entrant

	e.HasPillion = true
	e.IsBlank = true
	e.RiderName = "RIDER"
	e.PillionName = "PILLION"
	e.EventDate = CFG.EventDate
	return &e

}
func fileExists(x string) bool {

	_, err := os.Stat(x)
	return !errors.Is(err, os.ErrNotExist)

}

func loadConfig() {

	configPath := "rblrdox.yml"

	if !fileExists(configPath) {
		fmt.Printf("Can't find config file %v\n", configPath)
		return
	}

	file, err := os.Open(configPath)
	if err != nil {
		return
	}
	defer file.Close()

	D := yaml.NewDecoder(file)
	err = D.Decode(&CFG)
	if err != nil {
		panic(err)
	}

}
func main() {

	fmt.Printf("RBLRDOX v1.0\nCopyright (c) 2022 Bob Stammers\n")
	flag.Parse()
	loadConfig()
	fmt.Printf("Event date: %v\nGenerating %v into %v\n", CFG.EventDate, *doc, *outputfile)
	OUTF, _ = os.Create(*outputfile)
	defer OUTF.Close()
	var err error
	DBH, err = sql.Open("sqlite3", CFG.Database)
	if err != nil {
		panic(err)
	}
	defer DBH.Close()

	xfile := filepath.Join(*doc, "header.html")
	emitTopTail(OUTF, xfile)
	sql := "SELECT EntrantID,Bike,BikeReg,RiderName,RiderFirst,RiderIBA,PillionName,PillionFirst,PillionIBA"
	sql += ",OdoKms,Class,Phone,Email,NokName,NokRelation,NokPhone "
	sql += ",substr(RiderName,RiderPos+1) As RiderLast"
	sql += " FROM (SELECT *,instr(RiderName,' ') As RiderPos FROM entrants) "
	if *class != "" || *entrant != "" || *blanks > 0 {
		sql += " WHERE "
		if *blanks > 0 {
			sql += "EntrantID < 0 AND Class < 0" // So none will be found
		} else {
			if *class != "" {
				sql += "Class In (" + *class + ")"
				if *entrant != "" {
					sql += " OR " // Yes, or not and
				}
			}
			if *entrant != "" {
				sql += "EntrantID In (" + *entrant + ")"
			}
		}
	}
	sql += " ORDER BY RiderLast, RiderName" // Surname
	//fmt.Printf("%v\n", sql)
	rows, _ := DBH.Query(sql)
	NRex := 0
	for rows.Next() {
		e := newEntrant()
		e.IsBlank = false
		err := rows.Scan(&e.EntrantID, &e.Bike, &e.BikeReg, &e.RiderName, &e.RiderFirst, &e.RiderIBA,
			&e.PillionName, &e.PillionFirst, &e.PillionIBA,
			&e.OdoKms, &e.Class, &e.Phone, &e.Email, &e.NokName, &e.NokRelation, &e.NokPhone, &e.RiderLast)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		e.HasPillion = e.PillionName != ""

		xfile = filepath.Join(*doc, "entrant"+strconv.Itoa(e.Class)+".html")
		if !fileExists(xfile) {
			xfile = filepath.Join(*doc, "entrant.html")
		}
		if !fileExists(xfile) {
			fmt.Printf("Skipping Entrant %v %v; Class=%v\n", e.EntrantID, e.RiderName, e.Class)
			continue
		}
		t, err := template.ParseFiles(xfile)
		if err != nil {
			fmt.Printf("new %v\n", err)
		}
		err = t.Execute(OUTF, e)
		if err != nil {
			fmt.Printf("x %v\n", err)
		}
		NRex++
	}
	fmt.Printf("%v populated forms generated\n", NRex)
	if *blanks > 0 {
		printBlanks()
	}
	xfile = filepath.Join(*doc, "footer.html")
	emitTopTail(OUTF, xfile)
}

func emitTopTail(F *os.File, xfile string) {

	html, err := os.ReadFile(xfile)
	if err != nil {
		fmt.Printf("new %v\n", err)
	}
	F.WriteString(string(html))
}

func printBlanks() {

	classes := strings.Split(*class, ",")
	NREX := 0
	for _, c := range classes {
		for n := 0; n < *blanks; n++ {
			e := newEntrant()
			e.Class, _ = strconv.Atoi(c)
			e.HasPillion = !*solo
			xfile := filepath.Join(*doc, "entrant"+strconv.Itoa(e.Class)+".html")
			if !fileExists(xfile) {
				xfile = filepath.Join(*doc, "entrant.html")
			}

			t, err := template.ParseFiles(xfile)
			if err != nil {
				fmt.Printf("new %v\n", err)
			}
			err = t.Execute(OUTF, e)
			if err != nil {
				fmt.Printf("x %v\n", err)
			}
			NREX++
		}
	}
	fmt.Printf("%v blank forms generated\n", NREX)
}
