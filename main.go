package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/microsoft/go-mssqldb"
)

type Config struct {
	ServerName   string `json:"SERVER_NAME"`
	Port         int    `json:"PORT"`
	UserName     string `json:"USER_NAME"`
	UserPassword string `json:"USER_PASSWORD"`
	DatabaseName string `json:"DATABASE_NAME"`
}

func main() {
	var err error

	configJSON, err := os.Open("config.json")

	if err != nil {
		fmt.Println("Couldn't open the database config file: ", err.Error())
	}

	defer configJSON.Close()

	byteValue, _ := io.ReadAll(configJSON)

	var config Config

	json.Unmarshal(byteValue, &config)

	var db *sql.DB

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		config.ServerName, config.UserName, config.UserPassword, config.Port, config.DatabaseName)

	db, err = sql.Open("sqlserver", connString)

	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}

	ctx := context.Background()
	err = db.PingContext(ctx)

	if err != nil {
		log.Fatal(err.Error())
	}

	workItemSchema, err := os.ReadFile("schemas/work_item.sql")

	if err != nil {
		log.Fatal("Error reading schema file: ", err.Error())
	}

	_, err = db.ExecContext(ctx, string(workItemSchema))

	if err != nil {
		log.Fatal("Error executing schema SQL: ", err.Error())
	}

	addCmdCommand := flag.NewFlagSet("add", flag.ExitOnError)
	deleteCmdCommand := flag.NewFlagSet("delete", flag.ExitOnError)
	listCmdCommand := flag.NewFlagSet("list", flag.ExitOnError)

	workItemDate := addCmdCommand.String("date", "", "Work date (YYYY-MM-DD)")
	workItemStartTime := addCmdCommand.String("start", "", "Start time (HH:MM:SS)")
	workItemEndTime := addCmdCommand.String("end", "", "End time (HH:MM:SS)")
	workItemDescription := addCmdCommand.String("description", "", "Description of the work item")

	workItemID := deleteCmdCommand.String("id", "", "Id (n)")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'add', 'delete' or 'list' command")
	}

	switch os.Args[1] {
	case "add":
		addCmdCommand.Parse(os.Args[2:])
	case "delete":
		deleteCmdCommand.Parse(os.Args[2:])
	case "list":
		listCmdCommand.Parse(os.Args[2:])
	default:
		fmt.Println(fmt.Println("Expected 'add', 'delete' or 'list' command"))
		os.Exit(1)
	}

	if addCmdCommand.Parsed() {
		if *workItemDate == "" || *workItemStartTime == "" || *workItemEndTime == "" || *workItemDescription == "" {
			fmt.Println("All fields are required for adding a work item.")
			os.Exit(1)
		}

		_, err := db.ExecContext(ctx, "INSERT INTO WorkItem (workDate, startTime, endTime, description) VALUES (@p1, @p2, @p3, @p4)",
			*workItemDate, *workItemStartTime, *workItemEndTime, *workItemDescription)

		if err != nil {
			log.Fatal("Error adding work item: ", err.Error())
		}

		fmt.Println("Work item added successfully.")
	}

	if deleteCmdCommand.Parsed() {
		if *workItemID == "" {
			fmt.Println("Id is required for deleting a work item.")
			os.Exit(1)
		}

		_, err := db.ExecContext(ctx, "DELETE FROM WorkItem WHERE id = @p1", workItemID)

		if err != nil {
			log.Fatal("Error deleting work item: ", err.Error())
		}

		fmt.Println("Work item deleted successfully.")
	}

	if listCmdCommand.Parsed() {
		rows, err := db.QueryContext(ctx, "SELECT id, workDate, startTime, endTime, description FROM WorkItem")

		if err != nil {
			log.Fatal("Error retrieving work items: ", err.Error())
		}

		defer rows.Close()

		var workItems []WorkItem
		for rows.Next() {
			var item WorkItem

			if err := rows.Scan(&item.ID, &item.WorkDate, &item.StartTime, &item.EndTime, &item.Description); err != nil {
				log.Fatal(err)
			}

			workItems = append(workItems, item)
		}

		workItemsAsJSON, err := json.MarshalIndent(workItems, "", "  ")

		if err != nil {
			log.Fatal("Error converting work items to JSON: ", err.Error())
		}

		fmt.Println(string(workItemsAsJSON))
	}
}
