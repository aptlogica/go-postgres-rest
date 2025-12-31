package main

// import (
// 	"fmt"
// 	"godbgrest/internal/config"

// 	// "godbgrest/handlers"
// 	// "godbgrest/services"
// 	// "godbgrest/database"
// 	"godbgrest/internal/app"
// 	"log"
// )

// func main() {
// 	cfg, err := config.Load()
// 	if err != nil {
// 		log.Fatal("Failed to load config:", err)
// 	}

// 	// db, err := database.New(cfg.UserDatabase)
// 	// fmt.Println("Configuration loaded successfully:", cfg.UserDatabase)
// 	// fmt.Println("Configuration loaded successfully:", cfg.AppDatabase)

// 	// db, err := database.New(cfg.UserDatabase)
// 	if err != nil {
// 		fmt.Printf("failed to initialize database: %v", err)
// 	}

// 	// // Initialize repository
// 	// repo := database.NewRepository(db)
// 	// tableService := services.NewTableService(repo)
// 	// tableHandler := handlers.NewTableHandler(tableService)

// 	// tableHandler.GetTables()

// 	application, err := app.New(cfg)
// 	if err != nil {
// 		log.Fatal("Failed to create application:", err)
// 	}

// 	if err := application.Run(); err != nil {
// 		log.Fatal("Failed to run application:", err)
// 	}
// }

import (
	"context"
	"encoding/json"
	"godbgrest/pkg"
	"godbgrest/pkg/config"

	"fmt"
	godbgrestModels "godbgrest/pkg/models"
)

func prettierPrint(data any) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Failed to marshal data to JSON:", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

func createTable(repo *pkg.DatabaseService) {
	employeeTable := godbgrestModels.CreateTableRequest{
		Name: "employees",
		Columns: []godbgrestModels.ColumnDefinition{
			{Name: "id", DataType: "SERIAL", NotNull: true},
			{Name: "name", DataType: "VARCHAR(255)", NotNull: true},
			{Name: "email", DataType: "VARCHAR(255)", Unique: true},
			{Name: "department", DataType: "VARCHAR(100)"},
			{Name: "salary", DataType: "DECIMAL(10,2)"},
			{Name: "hire_date", DataType: "DATE", DefaultValue: func() *string { s := "CURRENT_DATE"; return &s }()},
			{Name: "active", DataType: "BOOLEAN", DefaultValue: func() *string { s := "true"; return &s }()},
		},
		PrimaryKey: []string{"id"},
		Indexes: []godbgrestModels.IndexDefinition{
			{Columns: []string{"email"}, Unique: true},
			{Columns: []string{"department"}},
		},
	}

	if err := repo.TableService.CreateTable(employeeTable); err != nil {
		fmt.Println(err)
	}
}

func addField(repo *pkg.DatabaseService) {
	addFieldReq := godbgrestModels.AddColumnRequest{
		Column: godbgrestModels.ColumnDefinition{
			Name: "gender", DataType: "VARCHAR(25)",
		},
	}
	if err := repo.TableService.AddColumn("employees", addFieldReq); err != nil {
		fmt.Println(err)
	}
}

func alterTableModifyField(repo *pkg.DatabaseService) {
	alterTableReq := godbgrestModels.AlterTableRequest{
		Action: "modify_column",
		Data: godbgrestModels.ModifyColumnRequest{
			ColumnName:  "gender",
			NewDataType: "VARCHAR(50)",
		},
	}
	if err := repo.TableService.AlterTable("employees", alterTableReq); err != nil {
		fmt.Println(err)
	}
}

func alterTableRenameField(repo *pkg.DatabaseService) {
	alterTableReq := godbgrestModels.AlterTableRequest{
		Action: "rename_column",
		Data: godbgrestModels.RenameColumnRequest{
			OldName: "gender",
			NewName: "employee_gender",
		},
	}
	if err := repo.TableService.AlterTable("employees", alterTableReq); err != nil {
		fmt.Println(err)
	}
}

func alterTableDropField(repo *pkg.DatabaseService) {
	alterTableReq := godbgrestModels.AlterTableRequest{
		Action: "drop_column",
		Data: godbgrestModels.DropColumnRequest{
			ColumnName: "employee_gender",
		},
	}
	if err := repo.TableService.AlterTable("employees", alterTableReq); err != nil {
		fmt.Println(err)
	}
}

func getTables(repo *pkg.DatabaseService) {

	data, err := repo.TableService.GetTables("public")
	if err != nil {
		fmt.Println(err)
	}
	prettierPrint(data)
}

func getTableData(repo *pkg.DatabaseService) {
	params := godbgrestModels.QueryParams{
		Select: []string{"*"},
	}

	data, err := repo.TableService.GetTableData(context.Background(), "employees", params)
	if err != nil {
		fmt.Println(err)
		return
	}
	prettierPrint(data)
}

func createRecord(repo *pkg.DatabaseService) {
	data := map[string]interface{}{
		"name":       "John Doe",
		"email":      "john@company.com",
		"department": "Engineering",
		"salary":     75000,
	}
	record, err := repo.TableService.CreateRecord(context.Background(), "employees", data)
	if err != nil {
		fmt.Println(err)
		return
	}
	prettierPrint(record)
}

func updateRecord(repo *pkg.DatabaseService) {
	data := map[string]interface{}{
		"department": "Finance",
	}
	updated, err := repo.TableService.UpdateRecord(context.Background(), "employees", 1, data)
	if err != nil {
		fmt.Println(err)
		return
	}
	prettierPrint(updated)
}

func deleteRecord(repo *pkg.DatabaseService, tableName string, id interface{}) {
	err := repo.TableService.DeleteRecord(context.Background(), tableName, id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Record deleted successfully")
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config:", err)
	}

	repo, err := pkg.NewDatabaseServiceWithInit(cfg)
	if err != nil {
		fmt.Println("Database Connection failed:", err)
	}

	// createTable(repo)

	// addField(repo)

	// alterTableModifyField(repo)
	// alterTableRenameField(repo)
	// alterTableDropField(repo)

	// getTables(repo)

	// createRecord(repo)
	updateRecord(repo)
	// getTableData(repo)
}
