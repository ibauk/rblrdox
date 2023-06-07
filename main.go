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
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const apptitle = "RBLRDOX v1.1"
const progdesc = `
I print disclaimers (-doc legal), receipt logs (-doc rlogs) and ride certificates (-doc certs)
for the RBLR1000 event using entrant details held in a ScoreMaster database.
The routes (class) are:- A-NC [2], B-NAC [1], C-SC [4], D-SAC [3], E-5C [6], F-5AC [7].
>24hr certs use class = A-NAC [8], B-NC [9], C-SAC [10], D-SC [11].
`

var db = flag.String("db", `\sm-installs\rblr23\sm\ScoreMaster.db`, "use this database")

var a5 = flag.Bool("a5", false, "Print two per A4 portrait page")
var doc = flag.String("doc", "rlogs", "The name of the document to be produced")
var solo = flag.Bool("solo", false, "Blank forms, rider only")
var class = flag.String("class", "", "The class numbers to be selected. Default=all")
var blanks = flag.Int("blanks", 0, "Print <n> blanks only")
var entrant = flag.String("entrant", "", "The entrant numbers to be selected. Default=all")
var rambling = flag.Bool("v", false, "Show debug info")
var showusage = flag.Bool("?", false, "Show this help")
var outputfile = flag.String("to", "output.html", "Output filename")

var DBH *sql.DB
var OUTF *os.File

var CFG struct {
	EventDate  string
	EventTitle string
	Database   string
	A5         bool
}

type Entrant struct {
	EntrantID    int
	ABike        string
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
	EventTitle   string
	PageAfter    bool
}

func newEntrant() *Entrant {

	var e Entrant

	e.HasPillion = true
	e.IsBlank = true
	e.RiderName = "RIDER"
	e.PillionName = "PILLION"
	e.EventDate = CFG.EventDate
	e.EventTitle = CFG.EventTitle
	e.PageAfter = true
	return &e

}
func fileExists(x string) bool {

	_, err := os.Stat(x)
	return !errors.Is(err, os.ErrNotExist)

}

func init() {

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%v\n", apptitle)
		fmt.Fprintf(w, "%v\n", progdesc)
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showusage {
		flag.Usage()
		os.Exit(1)
	}
	CFG.Database = *db
	CFG.A5 = *a5

}

// Crude English-only a/an prefix
func formattedABike(bike string) string {

	res := "a"
	switch bike[0:1] {
	case "A", "E", "I", "O":
		res = "an"
	}
	//fmt.Printf("%v == %v\n", bike, res)
	return res + " " + bike
}

func formattedDate(dt time.Time) string {

	return dt.Format("2 January 2006")

}

// daterange must be iso8601;iso8601
func formattedDateRange(daterange string) string {
	if daterange == "" {
		return ""
	}
	dts := strings.Split(daterange, ";")
	dt := strings.Split(dts[0], "T")
	dt1, _ := time.Parse("2006-01-02", dt[0])
	dt = strings.Split(dts[1], "T")
	dt2, _ := time.Parse("2006-01-02", dt[0])
	if dt1 == dt2 {
		return formattedDate(dt1)
	}
	if dt1.Year() == dt2.Year() && dt1.Month() == dt2.Month() {
		return dt1.Format("2") + "/" + formattedDate(dt2)
	}
	return formattedDate(dt1) + " - " + formattedDate(dt2)

}

func main() {

	fmt.Printf("%v\nCopyright (c) 2023 Bob Stammers\n", apptitle)
	fmt.Printf("%v\n", progdesc)
	OUTF, _ = os.Create(*outputfile)
	defer OUTF.Close()
	var err error
	DBH, err = sql.Open("sqlite3", CFG.Database)
	if err != nil {
		panic(err)
	}
	defer DBH.Close()

	sqlx := "SELECT rallyparams.StartTime || ';' || rallyparams.FinishTime AS DateRallyRange,RallyTitle FROM rallyparams"

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	if rows.Next() {
		var daterange string
		rows.Scan(&daterange, &CFG.EventTitle)
		CFG.EventDate = formattedDateRange(daterange)
	}
	rows.Close()
	fmt.Printf("Using database %v  Use -db to override\n", *db)
	fmt.Printf("Event date: %v\nGenerating %v into %v  Use -doc, -to to override\n", CFG.EventDate, *doc, *outputfile)
	xfile := filepath.Join(*doc, "header.html")
	if !fileExists(xfile) {
		fmt.Println(xfile + " doesn't exist, quitting")
		return
	}
	emitTopTail(OUTF, xfile)
	sqlx = "SELECT EntrantID,ifnull(Bike,''),ifnull(BikeReg,''),ifnull(RiderName,'') as RiderName,ifnull(RiderFirst,'')"
	sqlx += ",ifnull(RiderIBA,0),ifnull(PillionName,''),ifnull(PillionFirst,''),ifnull(PillionIBA,0)"
	sqlx += ",OdoKms,Class,ifnull(Phone,''),ifnull(Email,''),ifnull(NokName,''),ifnull(NokRelation,''),ifnull(NokPhone,'') "
	sqlx += ",substr(RiderName,RiderPos+1) As RiderLast"
	sqlx += " FROM (SELECT *,instr(RiderName,' ') As RiderPos FROM entrants) "
	if *class != "" || *entrant != "" || *blanks > 0 {
		sqlx += " WHERE "
		if *blanks > 0 {
			sqlx += "EntrantID < 0 AND Class < 0" // So none will be found
		} else {
			if *class != "" {
				sqlx += "Class In (" + *class + ")"
				if *entrant != "" {
					sqlx += " OR " // Yes, or not and
				}
			}
			if *entrant != "" {
				sqlx += "EntrantID In (" + *entrant + ")"
			}
		}
	}
	sqlx += " ORDER BY RiderLast, RiderName" // Surname
	if *rambling {
		fmt.Printf("%v\n", sqlx)
	}
	rows, err = DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
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
		e.ABike = formattedABike(e.Bike)
		if *a5 {
			e.PageAfter = NRex%2 != 0
		}
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

func emitParsedTopTail(F *os.File, xfile string) {
	t, err := template.ParseFiles(xfile)
	if err != nil {
		panic(err)
	}
	err = t.Execute(F, CFG)
	if err != nil {
		panic(err)
	}
}
func emitTopTail(F *os.File, xfile string) {

	emitParsedTopTail(F, xfile)

	/*
		html, err := os.ReadFile(xfile)
		if err != nil {
			fmt.Printf("new %v\n", err)
		}
		F.WriteString(string(html))
	*/
}

func printBlanks() {

	classes := strings.Split(*class, ",")
	NRex := 0
	for _, c := range classes {
		for n := 0; n < *blanks; n++ {
			e := newEntrant()
			e.Class, _ = strconv.Atoi(c)
			e.HasPillion = !*solo
			if *a5 {
				e.PageAfter = NRex%2 != 0
			}
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
			NRex++
		}
	}
	fmt.Printf("%v blank forms generated\n", NRex)
}
