package sheetsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

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
var BlueHighlight = sheets.Color{
	Red:   0.624,
	Green: 0.773,
	Blue:  0.91,
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

	Session = session
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

// SafeAccessor Provides a function to safely extract string values from the given Values object
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

func GetCurrentPage(svc *sheets.Service, sheetID string) (*sheets.Sheet, bool, error) {
	sheet, grantTime, removeTime, endTime, err := DoGetCurrentPage(svc, sheetID)
	if err != nil {
		return sheet, false, err
	}

	now := config.Now()
	fmt.Println("Start:", config.PrintTime(grantTime), "End:", config.PrintTime(endTime), config.PrintTime(now))
	fmt.Println("Found current page:", sheet.Properties.Title)
	return sheet, now.After(removeTime), nil
}

func DoGetCurrentPage(svc *sheets.Service, sheetID string) (*sheets.Sheet, time.Time, time.Time, time.Time, error) {
	resp, err := svc.Spreadsheets.Get(sheetID).Do()
	if err != nil {
		return &sheets.Sheet{}, time.Time{}, time.Time{}, time.Time{}, err
	}

	for _, sheet := range resp.Sheets {
		fmt.Println(sheet.Properties.Title, sheet.Properties.SheetId)
		r := fmt.Sprintf("'%s'!B1:B3", sheet.Properties.Title)
		resp2, err := svc.Spreadsheets.Values.Get(sheetID, r).Do()
		if err != nil {
			return &sheets.Sheet{}, time.Time{}, time.Time{}, time.Time{}, err
		}

		vals := SafeAccessor(resp2.Values)

		grantTime := config.ParseTime(vals(0, 0))
		removeTime := config.ParseTime(vals(1, 0))
		endTime := config.ParseTime(vals(2, 0))
		now := config.Now()

		if now.After(grantTime) && now.Before(endTime) {
			return sheet, grantTime, removeTime, endTime, nil
		}
	}
	return &sheets.Sheet{}, time.Time{}, time.Time{}, time.Time{}, errors.New("Not found")
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
	UserID        string
	TimeStr       string
	Plan          int
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
	ret := &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          r.Range.PageID,
				StartRowIndex:    r.Range.RowStart,
				EndRowIndex:      r.Range.RowEnd,
				StartColumnIndex: r.Range.ColStart,
				EndColumnIndex:   r.Range.ColEnd,
			},
			Fields: "userEnteredFormat.backgroundColor",
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: &color,
				},
			},
		},
	}

	return ret
}

func ReadAllAutomatic(svc *sheets.Service, sheetID string, sheet *sheets.Sheet) ([]RoleRow, error) {
	rows := []RoleRow{}

	r := fmt.Sprintf("'%s'!A7:D%d", sheet.Properties.Title, sheet.Properties.GridProperties.RowCount)
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
		plan, _ := strconv.Atoi(vals(i, 3))
		if plan == 0 {
			plan = 500
		}
		rowIndex := i + 7
		rr := RowRange{
			PageID:   sheet.Properties.SheetId,
			RowStart: int64(rowIndex) - 1,
			RowEnd:   int64(rowIndex),
			ColStart: 0,
			ColEnd:   4,
		}
		row := RoleRow{
			Row:           rowIndex,
			Range:         rr,
			Username:      username,
			Discriminator: disc,
			UserID:        vals(i, 1),
			TimeStr:       vals(i, 2),
			Plan:          plan,
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func ReadAllManual(svc *sheets.Service, sheetID string, sheet *sheets.Sheet) ([]RoleRow, error) {
	rows := []RoleRow{}

	r := fmt.Sprintf("'%s'!F7:K%d", sheet.Properties.Title, sheet.Properties.GridProperties.RowCount)
	resp, err := svc.Spreadsheets.Values.Get(sheetID, r).Do()
	if err != nil {
		log.Printf("Failed to read manual section, %s", err)
		return rows, err
	}
	vals := SafeAccessor(resp.Values)
	for i := 0; i < len(resp.Values); i++ {
		username, disc := ParseDiscordHandle(vals(i, 0))
		if username == "" {
			continue
		}
		plan, _ := strconv.Atoi(vals(i, 4))
		if plan == 0 {
			plan = 500
		}
		fmt.Println("READ MANUAL", username, disc, plan)
		rowIndex := i + 7
		rr := RowRange{
			PageID:   sheet.Properties.SheetId,
			RowStart: int64(rowIndex) - 1,
			RowEnd:   int64(rowIndex),
			ColStart: 5,
			ColEnd:   11,
		}
		row := RoleRow{
			Row:           rowIndex,
			Range:         rr,
			Username:      username,
			Discriminator: disc,
			UserID:        vals(i, 1),
			TimeStr:       vals(i, 2),
			Plan:          plan,
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func ReadAllExclude(svc *sheets.Service, sheetID string, sheet *sheets.Sheet) ([]RoleRow, error) {
	rows := []RoleRow{}

	r := fmt.Sprintf("'%s'!M7:Q%d", sheet.Properties.Title, sheet.Properties.GridProperties.RowCount)
	resp, err := svc.Spreadsheets.Values.Get(sheetID, r).Do()
	if err != nil {
		log.Printf("Failed to read excluded section, %s", err)
		return rows, err
	}
	vals := SafeAccessor(resp.Values)
	for i := 0; i < len(resp.Values); i++ {
		username, disc := ParseDiscordHandle(vals(i, 0))
		if username == "" || disc == "" {
			continue
		}
		rowIndex := i + 7
		rr := RowRange{
			PageID:   sheet.Properties.SheetId,
			RowStart: int64(rowIndex) - 1,
			RowEnd:   int64(rowIndex),
			ColStart: 12,
			ColEnd:   17,
		}
		row := RoleRow{
			Row:           rowIndex,
			Range:         rr,
			Username:      username,
			Discriminator: disc,
			UserID:        vals(i, 1),
			TimeStr:       vals(i, 2),
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func UpdateFormatting(svc *sheets.Service, sheetID string, reqs []*sheets.Request) error {
	req := sheets.BatchUpdateSpreadsheetRequest{
		Requests: reqs,
	}

	_, err := svc.Spreadsheets.BatchUpdate(sheetID, &req).Do()
	if err != nil {
		fmt.Println("ERROR", err)
		return err
	}

	return nil
}

func MapRows(rows []RoleRow) map[string]RoleRow {
	m := make(map[string]RoleRow)
	for _, row := range rows {
		m[row.UserID] = row
	}

	return m
}

func AddManualVerification(svc *sheets.Service, sheetID, handle, userID, proof string, plan int, verifiedBy string) error {
	page, _, err := GetCurrentPage(svc, sheetID)
	if err != nil {
		return err
	}

	r := fmt.Sprintf("'%s'!G7:G%d", page.Properties.Title, page.Properties.GridProperties.RowCount)
	resp, err := svc.Spreadsheets.Values.Get(sheetID, r).Do()
	if err != nil {
		return err
	}
	vals := SafeAccessor(resp.Values)

	fillRow := -1
	for i := 0; i < len(resp.Values); i++ {
		id := vals(i, 0)
		if id == "" || id == userID {
			fillRow = i + 7
		}
	}
	if fillRow == -1 {
		fillRow = len(resp.Values) + 7
	}
	r = fmt.Sprintf("'%s'!F%d:K%d", page.Properties.Title, fillRow, fillRow)
	rr := RowRange{
		PageID:   page.Properties.SheetId,
		RowStart: int64(fillRow) - 1,
		RowEnd:   int64(fillRow),
		ColStart: 5,
		ColEnd:   11,
	}
	row := RoleRow{Range: rr}
	colorReq := row.ColorRequest(GreenHighlight)

	if plan >= 10000 {
		colorReq = row.ColorRequest(YellowHighlight)
	} else if plan >= 1500 {
		colorReq = row.ColorRequest(BlueHighlight)
	}

	vr := &sheets.ValueRange{}
	vr.Values = append(vr.Values, []interface{}{handle, userID, config.PrintTime(config.Now()), proof, plan, verifiedBy})

	_, err = svc.Spreadsheets.Values.Update(sheetID, r, vr).ValueInputOption("RAW").Do()
	if err != nil {
		return err
	}

	err = UpdateFormatting(svc, sheetID, []*sheets.Request{colorReq})
	if err != nil {
		return err
	}

	return nil
}

func UpdateHandle(svc *sheets.Service, sheetID string, page *sheets.Sheet, row RoleRow, newHandle string) error {
	r := fmt.Sprintf("'%s'!F%d:F%d", page.Properties.Title, row.Row, row.Row)

	vr := &sheets.ValueRange{}
	vr.Values = append(vr.Values, []interface{}{newHandle})
	_, err := svc.Spreadsheets.Values.Update(sheetID, r, vr).ValueInputOption("RAW").Do()
	if err != nil {
		return err
	}
	return nil
}
