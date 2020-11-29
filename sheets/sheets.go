package sheets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/w8kerr/delubot/config"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var GreenHighlight = sheets.Color{
	Red:   0.714,
	Green: 0.843,
	Blue:  0.659,
	Alpha: 1.0,
}
var YellowHighlight = sheets.Color{
	Red:   1,
	Green: 0.898,
	Blue:  0.6,
	Alpha: 1.0,
}
var RedHighlight = sheets.Color{
	Red:   0.918,
	Green: 0.6,
	Blue:  0.6,
	Alpha: 1.0,
}

var GOOGLE_CLIENT_ID string
var GOOGLE_SECRET string
var Session *discordgo.Session

func Init(session *discordgo.Session) {
	GOOGLE_CLIENT_ID = os.Getenv("GOOGLE_CLIENT_ID")
	GOOGLE_SECRET = os.Getenv("GOOGLE_SECRET")
}

func GetService() (*sheets.Service, error) {
	credentialsJSON, err := json.Marshal(config.GoogleCredentials)
	if err != nil {
		log.Printf("Failed to form Google credentials, %s", err)
		return &sheets.Service{}, err
	}

	// Service account based oauth2 two legged integration
	ctx := context.Background()
	svc, err := sheets.NewService(ctx, option.WithCredentialsJSON(credentialsJSON), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		log.Printf("Failed to get Google Sheets service, %s", err)
		return svc, err
	}

	return svc, nil
}

// SafeVal Provides a function to safely extract string values from the given Values object
// (which is not necessarily padded for the full range)
func SafeAccessor(values [][]interface{}) func(int, int) string {
	return func(rowIndex, colIndex int) string {
		if len(values) > rowIndex {
			row := values[rowIndex]
			if len(row) > colIndex {
				cell := row[colIndex]
				str, ok := cell.(string)
				if ok {
					return str
				}
			}
		}

		return ""
	}
}

func GetCurrentPage(sheetID string) (*sheets.Sheet, error) {
	svc, err := GetService()
	if err != nil {
		return &sheets.Sheet{}, err
	}

	resp, err := svc.Spreadsheets.Get(sheetID).Do()
	if err != nil {
		return &sheets.Sheet{}, err
	}

	for _, sheet := range resp.Sheets {
		fmt.Println(sheet.Properties.Title, sheet.Properties.SheetId)
		r := fmt.Sprintf("'%s'!B1:B2", sheet.Properties.Title)
		resp2, err := svc.Spreadsheets.Values.Get(sheetID, r).Do()
		if err != nil {
			return &sheets.Sheet{}, err
		}

		vals := SafeAccessor(resp2.Values)

		startTime := config.ParseTime(vals(0, 0))
		endTime := config.ParseTime(vals(1, 0))
		now := config.Now()
		fmt.Println("Start:", config.PrintTime(startTime), "End:", config.PrintTime(endTime), config.PrintTime(now))
		if now.After(startTime) && now.Before(endTime) {
			fmt.Println("Found current page:", sheet.Properties.Title)
			return sheet, nil
		}
	}
	return &sheets.Sheet{}, errors.New("Not found")
}

func HasAccess(sheetID string) bool {
	svc, err := GetService()
	if err != nil {
		return false
	}

	_, err = svc.Spreadsheets.Values.Get(sheetID, "A1").Do()

	if err != nil {
		log.Printf("Failed to access sheet ID %s, %s", sheetID, err)
		return false
	}

	return true
}

var discordRE = regexp.MustCompile(`(.+)#(\d{4})$`)

func ParseDiscordHandle(handle string) (string, string) {
	matches := discordRE.FindAllSubmatch([]byte(handle), -1)
	if matches == nil {
		return handle, ""
	}

	return string(matches[0][1]), string(matches[0][2])
}

type RoleRow struct {
	Row           int
	Range         RowRange
	Username      string
	Discriminator string
	TimeStr       string
	Plan          string
}

type RowRange struct {
	PageID   int64
	RowStart int64
	RowEnd   int64
	ColStart int64
	ColEnd   int64
}

func (r *RoleRow) Handle() string {
	return fmt.Sprintf("%s#%s", r.Username, r.Discriminator)
}

func (r *RoleRow) ColorRequest(color sheets.Color) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          r.Range.PageID,
				StartRowIndex:    r.Range.RowStart,
				EndRowIndex:      r.Range.RowEnd,
				StartColumnIndex: r.Range.ColStart,
				EndColumnIndex:   r.Range.ColStart,
			},
			Fields: "userEnteredFormat.backgroundColor",
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: &color,
				},
			},
		},
	}
}

func ReadAllAutomatic(sheetID string, sheet *sheets.Sheet) ([]RoleRow, error) {
	rows := []RoleRow{}

	svc, err := GetService()
	if err != nil {
		log.Printf("Failed to get service to read automatic section, %s", err)
		return rows, err
	}

	r := fmt.Sprintf("'%s'!A6:C%d", sheet.Properties.Title, sheet.Properties.GridProperties.RowCount)
	resp, err := svc.Spreadsheets.Values.Get(sheetID, r).Do()
	if err != nil {
		log.Printf("Failed to read automatic section, %s", err)
		return rows, err
	}
	vals := SafeAccessor(resp.Values)
	for i := 0; i < len(resp.Values); i++ {
		username, disc := ParseDiscordHandle(vals(i, 0))
		if username == "" {
			continue
		}
		rowIndex := i + 6
		rr := RowRange{
			PageID:   sheet.Properties.SheetId,
			RowStart: int64(rowIndex) - 1,
			RowEnd:   int64(rowIndex),
			ColStart: 0,
			ColEnd:   3,
		}
		row := RoleRow{
			Row:           rowIndex,
			Range:         rr,
			Username:      username,
			Discriminator: disc,
			TimeStr:       vals(i, 1),
			Plan:          vals(i, 2),
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func ReadAllManual(sheetID string, sheet *sheets.Sheet) ([]RoleRow, error) {
	rows := []RoleRow{}

	svc, err := GetService()
	if err != nil {
		log.Printf("Failed to get service to read automatic section, %s", err)
		return rows, err
	}

	r := fmt.Sprintf("'%s'!E6:H%d", sheet.Properties.Title, sheet.Properties.GridProperties.RowCount)
	resp, err := svc.Spreadsheets.Values.Get(sheetID, r).Do()
	if err != nil {
		log.Printf("Failed to read automatic section, %s", err)
		return rows, err
	}
	vals := SafeAccessor(resp.Values)
	for i := 0; i < len(resp.Values); i++ {
		username, disc := ParseDiscordHandle(vals(i, 0))
		if username == "" {
			continue
		}
		rowIndex := i + 6
		rr := RowRange{
			PageID:   sheet.Properties.SheetId,
			RowStart: int64(rowIndex) - 1,
			RowEnd:   int64(rowIndex),
			ColStart: 4,
			ColEnd:   9,
		}
		row := RoleRow{
			Row:           rowIndex,
			Range:         rr,
			Username:      username,
			Discriminator: disc,
			TimeStr:       vals(i, 1),
			Plan:          vals(i, 3),
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func ReadAllExclude(sheetID string, sheet *sheets.Sheet) ([]RoleRow, error) {
	rows := []RoleRow{}

	svc, err := GetService()
	if err != nil {
		log.Printf("Failed to get service to read automatic section, %s", err)
		return rows, err
	}

	r := fmt.Sprintf("'%s'!K6:L%d", sheet.Properties.Title, sheet.Properties.GridProperties.RowCount)
	resp, err := svc.Spreadsheets.Values.Get(sheetID, r).Do()
	if err != nil {
		log.Printf("Failed to read automatic section, %s", err)
		return rows, err
	}
	vals := SafeAccessor(resp.Values)
	for i := 0; i < len(resp.Values); i++ {
		username, disc := ParseDiscordHandle(vals(i, 0))
		if username == "" || disc == "" {
			continue
		}
		rowIndex := i + 6
		rr := RowRange{
			PageID:   sheet.Properties.SheetId,
			RowStart: int64(rowIndex) - 1,
			RowEnd:   int64(rowIndex),
			ColStart: 0,
			ColEnd:   3,
		}
		row := RoleRow{
			Row:           rowIndex,
			Range:         rr,
			Username:      username,
			Discriminator: disc,
			TimeStr:       vals(i, 1),
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func MapRows(rows []RoleRow) map[string]RoleRow {
	m := make(map[string]RoleRow)
	for _, row := range rows {
		m[row.Handle()] = row
	}

	return m
}
