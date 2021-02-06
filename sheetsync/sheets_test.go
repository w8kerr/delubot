package sheetsync

import (
	"fmt"
	"log"
	"testing"

	"github.com/w8kerr/delubot/config"
	"github.com/w8kerr/delubot/mongo"
	"github.com/w8kerr/delubot/utils"
	"google.golang.org/api/sheets/v4"
)

func Test_Read(t *testing.T) {
	// guildID := "782092598290546719"
	// fmt.Println(guildID)
	sheetID := "1P0LxIbOAD995a5gceLLeAUhC5faWFmldF8CfhaQVocI"
	mongo.Init(true)
	config.LoadConfig()

	svc, err := GetService()
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}

	resp, err := svc.Spreadsheets.Get(sheetID).Do()
	for _, sheet := range resp.Sheets {
		fmt.Println(sheet.Properties.Title, sheet.Properties.SheetId)
		r := fmt.Sprintf("'%s'!A6:C%d", sheet.Properties.Title, sheet.Properties.GridProperties.RowCount)
		fmt.Println("RANGE", r)
		resp2, err := svc.Spreadsheets.Values.Get(sheetID, r).Do()
		fmt.Println("HAS ACCESS READ", err)
		utils.PrintJSON(resp2)
	}
}

func Test_GetCurrentPage(t *testing.T) {
	sheetID := "1P0LxIbOAD995a5gceLLeAUhC5faWFmldF8CfhaQVocI"
	mongo.Init(true)
	config.LoadConfig()

	svc, err := GetService()
	if err != nil {
		log.Printf("Couldn't create Sheet service, %s", err)
		return
	}

	GetCurrentPage(svc, sheetID)
}

func Test_DiscordRegex(t *testing.T) {
	username, disc := ParseDiscordHandle("Mirr#8388#8388")
	fmt.Println("Username:", username)
	fmt.Println("Discriminator:", disc)
}

func Test_ReadAllAutomatic(t *testing.T) {
	sheetID := "1P0LxIbOAD995a5gceLLeAUhC5faWFmldF8CfhaQVocI"
	mongo.Init(true)
	config.LoadConfig()

	svc, err := GetService()
	if err != nil {
		log.Printf("Couldn't create Sheet service, %s", err)
		return
	}

	page, _, _, err := GetCurrentPage(svc, sheetID)
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}

	rows, err := ReadAllAutomatic(svc, sheetID, page)
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}

	fmt.Println("ROWS")
	utils.PrintJSON(rows)
}

func Test_Formatting(t *testing.T) {
	sheetID := "1P0LxIbOAD995a5gceLLeAUhC5faWFmldF8CfhaQVocI"
	mongo.Init(true)
	config.LoadConfig()

	svc, err := GetService()
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}

	req := sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				RepeatCell: &sheets.RepeatCellRequest{
					Range: &sheets.GridRange{
						SheetId:          0,
						StartRowIndex:    5,
						EndRowIndex:      6,
						StartColumnIndex: 0,
						EndColumnIndex:   3,
					},
					Fields: "userEnteredFormat.backgroundColor",
					Cell: &sheets.CellData{
						UserEnteredFormat: &sheets.CellFormat{
							BackgroundColor: &GreenHighlight,
						},
					},
				},
			},
		},
	}

	resp, err := svc.Spreadsheets.BatchUpdate(sheetID, &req).Do()
	if err != nil {
		fmt.Println("ERROR", err)
	}

	utils.PrintJSON(resp)
}
