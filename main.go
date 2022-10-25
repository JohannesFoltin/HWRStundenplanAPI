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
	linkToData   string
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Serveradress and Port (Format: xxx.xxx.xxx.xxx:xxxx): ")
	serverAdress, _ := reader.ReadString('\n')
	//unnötiges Gedöhns. Leider funktioniert es sonst nicht
	serverAdress = strings.Replace(serverAdress, "\n", "", -1)
	linkToData = strings.Replace(linkToData, "\r", "", -1)
	fmt.Println(serverAdress)
	
	readerlink := bufio.NewReader(os.Stdin)
	fmt.Print("Aus Sicherheitsgründen, bitte gebe den Link zur ICS datei von deinem Kurs an: ")
	linkToData, _ = readerlink.ReadString('\n')
	linkToData = strings.Replace(linkToData, "\n", "", -1)
	linkToData = strings.Replace(linkToData, "\r", "", -1)
	fmt.Println(linkToData)
	
	router := gin.Default()

	router.GET("/plan", getPlan)

	router.Run(serverAdress)
}

func  getData() (Plan) {
	//Zieh mir den Quatsch aus dem Internet
	lastStundenplanData,err := http.Get(linkToData)
	if err != nil {
		fmt.Println(err)
	}
	defer lastStundenplanData.Body.Close()

	//Parse den Quatsch
	end := time.Now().Add(12*30*24*time.Hour)
	c := gocal.NewParser(lastStundenplanData.Body)
	c.End = &end
	c.Parse()

	plani := Plan{make([]Vorlesung, 0)}

	//Frag nicht. Macht es einigermaßen "huebsch"
	for _, e := range c.Events {
		startZeit, endZeit := e.Start.UnixMilli(), e.End.UnixMilli()
		vorlesung := Vorlesung{startZeit, endZeit, e.Location, SummaryParser(e.Summary)}
		if vorlesung.Raum == "" {
			if strings.Contains(e.Summary, "Klausur") {
				vorlesung.Raum = "Klausur"
			} else {
				vorlesung.Raum = "Online"
			}
		}
		plani.Vorlesungen = append(plani.Vorlesungen, vorlesung)
	}
	fmt.Print(len(plani.Vorlesungen), " Data received")
	return plani
}

func getPlan(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, getData())
}

func SummaryParser(summery string) string {
	lvl1 := strings.ReplaceAll(summery, ";", " ")
	lvl2 := strings.Split(lvl1, "CL:")
	lvl3 := strings.Split(string(lvl2[0]), "ONL")
	return lvl3[0]
}
