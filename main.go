package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
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
	lastStundenplanData http.Response
	linkToData string
)

func main() {
	
	getData()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Serveradress and Port (Format: xxx.xxx.xxx.xxx:xxxx): ")
	serverAdress, _ := reader.ReadString('\n')

	readerlink := bufio.NewReader(os.Stdin)
	fmt.Print("Aus Sicherheitsgründen, bitte gebe den Link zur ICS datei von deinem Kurs an: ")
	linkToData, _ = readerlink.ReadString('\n')

	router := gin.Default()

	router.GET("/plan", getPlan)
	router.POST("/plan", updateData)

	router.Run(serverAdress)
}
func updateData(d *gin.Context){

	f, err := http.Get(linkToData)

	if err != nil {
			fmt.Println(err)
	}
	defer f.Body.Close()

	if (f.Body != lastStundenplanData.Body){
		getData()
	}

}
func getData(){
	//Zieh mir den Quatsch aus dem Internet
	plan = Plan{make([]Vorlesung, 0)}
	lastStundenplanData, err := http.Get(linkToData)
	if err != nil {
			fmt.Println(err)
	}
	defer lastStundenplanData.Body.Close()

	//Parse den Quatsch
	start, end := time.Now(), time.Now().Add(12*30*24*time.Hour)
	c := gocal.NewParser(lastStundenplanData.Body)
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
	lvl1 := strings.ReplaceAll(summery, ";", " ")
	lvl2 := strings.Split(lvl1, "CL:")
	lvl3 := strings.Split(string(lvl2[0]), "ONL")
	return lvl3[0]
}
