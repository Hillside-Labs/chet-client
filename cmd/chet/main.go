package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hillside-labs/chet-client/models"
	"github.com/hillside-labs/chet-client/lib"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
)

func defaultDB() string {
	homedir, _ := os.UserHomeDir()
	return filepath.Join(homedir, ".chet.db")
}

func SaveLocalConfigFromFlags(c *cli.Context) error {
	db, err := models.Connect(defaultDB())
	if err != nil {
		fmt.Println("chet: couldn't open ~/.chet.db")
		return err
	}

	config := NewConfigFromFlags(c)

	// We only need one row
	config.ID = 1

	if txn := db.Save(&config); txn.Error != nil {
		return txn.Error
	}

	fmt.Println("Config saved!")
	return nil
}

func NewConfigFromFlags(c *cli.Context) models.LocalConfig {
	return models.LocalConfig{
		UserEmail:     c.String("email"),
		ClientID:      c.String("client-id"),
		ClientSecret:  c.String("client-secret"),
		ServerAddress: c.String("addr"),
		DisableRemote: c.Bool("disable-remote-reporting"),
	}
}

func WithConfigFlags(flags []cli.Flag) []cli.Flag {
	return append(flags,
		&cli.StringFlag{
			Name:    "addr",
			Aliases: []string{"a"},
			Value:   "http://localhost:4001",
		},
		&cli.StringFlag{
			Name:    "client-id",
			Aliases: []string{"c"},
		},
		&cli.StringFlag{
			Name:    "client-secret",
			Aliases: []string{"s"},
		},
		&cli.StringFlag{
			Name:    "email",
			Aliases: []string{"e"},
		},
		&cli.StringFlag{
			Name:    "disable-remote-reporting",
			Aliases: []string{"d"},
		},
	)
}

func main() {
	var outputFlag = &cli.StringFlag{
		Name:        "output",
		Aliases:     []string{"o", "format"},
		Value:       "table",
		Usage:       "Output format. One of: table, csv, markdown, json",
		DefaultText: "table",
	}

	app := cli.App{
		Name:  "chet",
		Usage: "Track command times.",
		Flags: WithConfigFlags([]cli.Flag{}),
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				cli.ShowAppHelpAndExit(c, 1)
			}
			cmd := exec.Command(c.Args().First(), c.Args().Tail()...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			start := time.Now()
			err := cmd.Run()
			total := time.Since(start)

			if err := Track(cmd, total, NewConfigFromFlags(c)); err != nil {
				fmt.Println(err)
			}

			return err
		},
		Commands: []*cli.Command{
			{
				Name:  "config",
				Flags: WithConfigFlags([]cli.Flag{}),
				Action: func(c *cli.Context) error {
					return SaveLocalConfigFromFlags(c)
				},
			},
			{
				Name:  "auth",
				Flags: WithConfigFlags([]cli.Flag{}),
				Action: func(c *cli.Context) error {
					err := SaveLocalConfigFromFlags(c)
					if err != nil {
						return err
					}
					db, err := models.Connect(defaultDB())
					if err != nil {
						fmt.Println("chet: couldn't open ~/.chet.db")
					}

					var config models.LocalConfig
					err = db.Last(&config).Error
					if err != nil {
						fmt.Println("failed to retrieve config")
					}

					_, err = GetTokenFromConfig(&config)
					if err != nil {
						return err
					}

					return db.Save(config).Error
				},
			},
			{
				Name: "report",
				Flags: []cli.Flag{
					outputFlag,
				},
				Action: func(c *cli.Context) error {
					db, err := models.Connect(defaultDB())
					if err != nil {
						return err
					}
					type result struct {
						Label      string
						TotalCalls int
						TotalTime  int
					}
					results := []result{}

					db.Model(&models.Record{}).
						Select("label, count(*) as total_calls , sum(duration) as total_time").
						Group("label").Find(&results)
					rows := []table.Row{}
					for _, r := range results {
						rows = append(rows, table.Row{r.Label, r.TotalCalls, time.Duration(r.TotalTime)})
					}

					err = OutputReport(c.String("output"), []string{"Label", "Total Calls", "Total Time"}, rows)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name: "query",
				Flags: []cli.Flag{
					outputFlag,
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("please provide a gmail-style query")
					}

					jsonQuery, err := lib.NewJSONQueryFromQuery(c.Args().First())
					if err != nil {
						return err
					}

					jsonQuery.Table = "records"

					db, err := models.Connect(defaultDB())
					if err != nil {
						return err
					}

					records, err := jsonQuery.Query(db)
					if err != nil {
						return err
					}

					columns := []string{"created_at", "username", "cmd", "duration", "repo", "branch", "os", "container"}
					rows := []table.Row{}
					for _, r := range records {
						rows = append(rows, table.Row{
							r.CreatedAt.Format("2006/1/2 3:04pm"),
							r.Username,
							r.Cmd,
							r.Duration.String(),
							r.Repo,
							r.Branch,
							r.OS,
							fmt.Sprintf("%v", r.Container),
						})
					}
					err = OutputReport(c.String("output"), columns, rows)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:  "sql",
				Usage: "Query the local sqlite db via sql",
				Flags: []cli.Flag{
					outputFlag,
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("please provide a SQL query")
					}

					db, err := models.Connect(defaultDB())
					if err != nil {
						return err
					}

					results, err := db.Raw(c.Args().First()).Rows()
					if err != nil {
						return err
					}

					defer results.Close()

					cols, err := results.Columns()
					if err != nil {
						return err
					}

					var rows []table.Row

					// Create a slice of interface{}'s to hold the values
					values := make([]interface{}, len(cols))
					for i := range values {
						values[i] = new(interface{})
					}

					// Loop through the rows
					for results.Next() {
						// Scan the values into the interface{} slice
						err := results.Scan(values...)
						if err != nil {
							return err
						}

						// Create a new row and append it to the rows slice
						row := make(table.Row, len(cols))
						for i, v := range values {
							row[i] = fmt.Sprintf("%v", *v.(*interface{}))
						}
						rows = append(rows, row)
					}

					err = OutputReport(c.String("output"), cols, rows)
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func OutputReport(format string, columns []string, rows []table.Row) error {
	if format == "json" {
		return OutputJSON(columns, rows)
	}
	return OutputTable(format, columns, rows)
}

func OutputTable(format string, columns []string, rows []table.Row) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	headers := table.Row{}
	for _, col := range columns {
		headers = append(headers, col)
	}
	t.AppendHeader(headers)
	t.AppendRows(rows)
	t.SetStyle(table.StyleLight)
	switch format {
	case "table":
		t.Render()
	case "csv":
		t.RenderCSV()
	case "markdown":
		t.RenderMarkdown()
	}
	return nil
}

type Row struct {
	Data map[string]interface{} `json:"data"`
}

func (r Row) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Data)
}

func OutputJSON(columns []string, rows []table.Row) error {
	jsonRows := make([]Row, len(rows))
	for i, row := range rows {
		jsonRow := Row{
			Data: make(map[string]interface{}),
		}

		for j, cell := range row {
			jsonRow.Data[columns[j]] = cell
		}

		jsonRows[i] = jsonRow
	}

	jsonBytes, err := json.Marshal(jsonRows)
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))

	return nil
}
