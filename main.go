package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/apognu/gocal"
	"github.com/gin-gonic/gin"
)

type Plan struct {
	Vorlesungen []Vorlesung
}

type Vorlesung struct {
	Startzeit    int64
	Endzeit      int64
	Raum         string
	Beschreibung string
}

var (
	plan Plan
)

func main() {
	dataManager(nil)
	router := gin.Default()

	router.GET("/plan", getPlan)
	//Sehr hässliche Lösung. Soll aber einfach nur die Daten neu ziehen und nichts aus der GET Message ziehen
	router.POST("/plan", dataManager)

	router.Run("localhost:3333")
}
func dataManager(d *gin.Context){
	//Zieh mir den Quatsch aus dem Internet
	plan = Plan{make([]Vorlesung, 0)}
	f, err := http.Get("https://moodle.hwr-berlin.de/fb2-stundenplan/download.php?doctype=.ics&url=./fb2-stundenplaene/informatik/semester1/kursb")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Body.Close()

	//Parse den Quatsch
	start, end := time.Now(), time.Now().Add(12*30*24*time.Hour)
	c := gocal.NewParser(f.Body)
	c.Start, c.End = &start, &end
	c.Parse()

	//Frag nicht. Macht es einigermaßen "huebsch"
	for _, e := range c.Events {
		startZeit, endZeit := e.Start.UnixMilli(), e.End.UnixMilli()
		vorlesung := Vorlesung{startZeit, endZeit, e.Location, SummaryParser(e.Summary)}
		if vorlesung.Raum == "" {
			vorlesung.Raum = "Online"
		}
		plan.Vorlesungen = append(plan.Vorlesungen, vorlesung)
	}
	fmt.Print(len(plan.Vorlesungen)," Update")

}

func getPlan(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, plan)
}

func SummaryParser(summery string) string {
	//ändert alle ; auf \n für Absätze
	lvl1 := strings.ReplaceAll(summery, ";", "\n")
	//Löscht die Räume aus der Summary, da sie abesondert in Room stehen und deswegen doppelt wären
	lvl2 := strings.Split(lvl1, "CL:")
	//Gleiche wie oben nur für Online Unterricht
	lvl3 := strings.Split(string(lvl2[0]), "ONL")
	return lvl3[0]
}
