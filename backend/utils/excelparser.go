package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Driver struct {
	Name          string
	ShiftHours    int
	PastWeekHours []int
}

type Route struct {
	RouteID     int
	DistanceKm  float64
	Traffic     string
	BaseTimeMin int
}

type Order struct {
	OrderID      int
	ValueRs      float64
	RouteID      int
	DeliveryTime string
}

// Parse Drivers from Excel
func LoadDrivers(filePath string) ([]Driver, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}
	var drivers []Driver
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		shiftHours, _ := strconv.Atoi(row[1])
		pastWeekStr := strings.Split(row[2], "|")
		var pastWeek []int
		for _, v := range pastWeekStr {
			val, _ := strconv.Atoi(v)
			pastWeek = append(pastWeek, val)
		}
		drivers = append(drivers, Driver{
			Name:          row[0],
			ShiftHours:    shiftHours,
			PastWeekHours: pastWeek,
		})
	}
	return drivers, nil
}

// Parse Routes from Excel
func LoadRoutes(filePath string) ([]Route, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}
	var routes []Route
	for i, row := range rows {
		if i == 0 {
			continue
		}
		id, _ := strconv.Atoi(row[0])
		dist, _ := strconv.ParseFloat(row[1], 64)
		baseTime, _ := strconv.Atoi(row[3])
		routes = append(routes, Route{
			RouteID:     id,
			DistanceKm:  dist,
			Traffic:     row[2],
			BaseTimeMin: baseTime,
		})
	}
	return routes, nil
}

// Parse Orders from Excel
func LoadOrders(filePath string) ([]Order, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}
	var orders []Order
	for i, row := range rows {
		if i == 0 {
			continue
		}
		id, _ := strconv.Atoi(row[0])
		val, _ := strconv.ParseFloat(row[1], 64)
		routeID, _ := strconv.Atoi(row[2])
		orders = append(orders, Order{
			OrderID:      id,
			ValueRs:      val,
			RouteID:      routeID,
			DeliveryTime: row[3],
		})
	}
	return orders, nil
}

func TestExcelParsers() {
	drivers, err := LoadDrivers("data/drivers.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Drivers:", drivers)

	routes, _ := LoadRoutes("data/routes.xlsx")
	fmt.Println("Routes:", routes)

	orders, _ := LoadOrders("data/orders.xlsx")
	fmt.Println("Orders:", orders)
}
func TestLoadAll() {
	drivers, _ := LoadDrivers("data/drivers.xlsx")
	routes, _ := LoadRoutes("data/routes.xlsx")
	orders, _ := LoadOrders("data/orders.xlsx")

	fmt.Println("Drivers:", drivers)
	fmt.Println("Routes:", routes)
	fmt.Println("Orders:", orders)
}
func ToInterfaceSlice[T any](slice []T) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}
