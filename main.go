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

	_ "github.com/mattn/go-sqlite3"
)

var doc = flag.String("doc", "rlogs", "The name of the document to be produced")
var class = flag.Int("class", -1, "The class number to be selected. Default=all")
var entrant = flag.Int("entrant", -1, "The entrant number to be selected. Default=all")
var outputfile = flag.String("to", "output.html", "Output filename")

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
}

func fileExists(x string) bool {

	_, err := os.Stat(x)
	return !errors.Is(err, os.ErrNotExist)

}

func main() {

	fmt.Println("RBLRDOX v0.1")
	flag.Parse()
	fmt.Printf("Generating %v\n", *doc)
	F, _ := os.Create(*outputfile)
	defer F.Close()
	DBH, err := sql.Open("sqlite3", "scoremaster.db")
	if err != nil {
		panic(err)
	}
	defer DBH.Close()

	xfile := filepath.Join(*doc, "header.html")
	emitTopTail(F, xfile)
	sql := "SELECT EntrantID,Bike,BikeReg,RiderName,RiderFirst,RiderIBA,PillionName,PillionFirst,PillionIBA"
	sql += ",OdoKms,Class,Phone,Email,NokName,NokRelation,NokPhone "
	sql += ",substr(RiderName,RiderPos+1) As RiderLast"
	sql += " FROM (SELECT *,instr(RiderName,' ') As RiderPos FROM entrants) "
	if *class >= 0 || *entrant >= 0 {
		sql += " WHERE "
		if *class >= 0 {
			sql += "Class=" + strconv.Itoa(*class)
			if *entrant >= 0 {
				sql += " OR " // Yes, or not and
			}
		}
		if *entrant >= 0 {
			sql += "EntrantID=" + strconv.Itoa(*entrant)
		}
	}
	sql += " ORDER BY RiderLast, RiderName" // Surname
	//fmt.Printf("%v\n", sql)
	rows, _ := DBH.Query(sql)
	for rows.Next() {
		var e Entrant
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
		err = t.Execute(F, e)
		if err != nil {
			fmt.Printf("x %v\n", err)
		}

	}
	xfile = filepath.Join(*doc, "footer.html")
	emitTopTail(F, xfile)
}

func emitTopTail(F *os.File, xfile string) {

	html, err := os.ReadFile(xfile)
	if err != nil {
		fmt.Printf("new %v\n", err)
	}
	F.WriteString(string(html))
}
