// package handlers

// import (
// 	"context"
// 	"fmt"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/xuri/excelize/v2"

// 	"github.com/parshwanath-p2493/GreenCartt/config"
// 	"github.com/parshwanath-p2493/GreenCartt/models"
// )

// // SeedFromExcel handles uploaded .xlsx with sheets: "orders", "drivers", "routes".
// // It accepts the formats you provided: '|' delimited past_week_hours, HH:MM delivery_time (no date).
// func SeedFromExcel(c *fiber.Ctx) error {
// 	// expects file upload via form field "file"
// 	fileHeader, err := c.FormFile("file")
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "file required", "details": err.Error()})
// 	}
// 	f, err := fileHeader.Open()
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "cannot open file", "details": err.Error()})
// 	}
// 	defer f.Close()

// 	// read into excelize
// 	xl, err := excelize.OpenReader(f)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "invalid excel", "details": err.Error()})
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
// 	defer cancel()

// 	// optional: clear existing collections so seeding is idempotent
// 	_ = config.DB.Collection("drivers").Drop(ctx)
// 	_ = config.DB.Collection("routes").Drop(ctx)
// 	_ = config.DB.Collection("orders").Drop(ctx)

// 	// -------------------
// 	// PARSE ROUTES sheet
// 	// expected header: route_id,distance_km,traffic_level,base_time_min
// 	if rows, err := xl.GetRows("routes"); err == nil && len(rows) > 0 {
// 		var routes []interface{}
// 		for i, r := range rows {
// 			if i == 0 {
// 				continue // skip header
// 			}
// 			// guard
// 			if len(r) < 4 {
// 				continue
// 			}
// 			routeID := strings.TrimSpace(r[0])
// 			// normalize routeID to string (keep as string)
// 			distance := ParseFloatOrZero(r[1])
// 			traffic := strings.TrimSpace(r[2])
// 			baseMin := ParseFloatOrZero(r[3])

// 			route := models.Route{
// 				RouteID:      routeID,
// 				DistanceKm:   distance,
// 				TrafficLevel: traffic,
// 				BaseTimeMin:  baseMin,
// 			}
// 			routes = append(routes, route)
// 		}
// 		if len(routes) > 0 {
// 			if _, err := config.DB.Collection("routes").InsertMany(ctx, routes); err != nil {
// 				return c.Status(500).JSON(fiber.Map{"error": "routes insert failed", "details": err.Error()})
// 			}
// 		}
// 	}

// 	// -------------------
// 	// PARSE DRIVERS sheet
// 	// expected header: name,shift_hours,past_week_hours (past_week_hours uses '|' as separator)
// 	if rows, err := xl.GetRows("drivers"); err == nil && len(rows) > 0 {
// 		var drivers []interface{}
// 		for i, r := range rows {
// 			if i == 0 {
// 				continue
// 			} // header
// 			if len(r) < 3 {
// 				continue
// 			}
// 			name := strings.TrimSpace(r[0])
// 			shiftHrs := ParseFloatOrZero(r[1]) // shift_hours might be numeric
// 			rawPast := strings.TrimSpace(r[2]) // like: 6|8|7|7|7|6|10

// 			parts := SplitAndTrimAny(rawPast, []string{"|", ","})
// 			past := []int{}
// 			for _, p := range parts {
// 				if p == "" {
// 					continue
// 				}
// 				v, err := strconv.Atoi(p)
// 				if err == nil {
// 					past = append(past, v)
// 				}
// 			}
// 			// create an ID for driver (safe, deterministic): lower(name) without spaces
// 			id := MakeDriverID(name)

// 			d := models.Driver{
// 				ID:              id,
// 				Name:            name,
// 				CurrentShiftHrs: shiftHrs,
// 				Past7DayHours:   past,
// 			}
// 			drivers = append(drivers, d)
// 		}
// 		if len(drivers) > 0 {
// 			if _, err := config.DB.Collection("drivers").InsertMany(ctx, drivers); err != nil {
// 				return c.Status(500).JSON(fiber.Map{"error": "drivers insert failed", "details": err.Error()})
// 			}
// 		}
// 	}

// 	// -------------------
// 	// PARSE ORDERS sheet
// 	// expected header: order_id,value_rs,route_id,delivery_time (delivery_time is HH:MM)
// 	if rows, err := xl.GetRows("orders"); err == nil && len(rows) > 0 {
// 		var orders []interface{}
// 		for i, r := range rows {
// 			if i == 0 {
// 				continue
// 			}
// 			if len(r) < 4 {
// 				continue
// 			}
// 			orderID := strings.TrimSpace(r[0])
// 			value := ParseFloatOrZero(r[1])
// 			routeID := strings.TrimSpace(r[2]) // may be numeric -> keep as string
// 			dtime := strings.TrimSpace(r[3])   // "02:07"

// 			// parse HH:MM into a timestamp: attach today's date (UTC) or a fixed date (2025-08-12)
// 			// We'll attach today's date (UTC) for consistency
// 			ts, perr := ParseHHMMToTimeUTC(dtime)
// 			if perr != nil {
// 				// fallback: zero time
// 				ts = time.Now().UTC()
// 			}

// 			o := models.Order{
// 				OrderID:        orderID,
// 				ValueRs:        value,
// 				AssignedRoute:  routeID, // matches models field assigned_route_id json tag
// 				DeliveryTimeTs: ts,
// 			}
// 			orders = append(orders, o)
// 		}
// 		if len(orders) > 0 {
// 			if _, err := config.DB.Collection("orders").InsertMany(ctx, orders); err != nil {
// 				return c.Status(500).JSON(fiber.Map{"error": "orders insert failed", "details": err.Error()})
// 			}
// 		}
// 	}

// 	return c.JSON(fiber.Map{"status": "ok", "message": "seed completed"})
// }

// // helper: make driver id from name
// func MakeDriverID(name string) string {
// 	s := strings.ToLower(strings.TrimSpace(name))
// 	s = strings.ReplaceAll(s, " ", "-")
// 	// remove any non-alphanumeric or dash
// 	out := []rune{}
// 	for _, r := range s {
// 		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
// 			out = append(out, r)
// 		}
// 	}
// 	return string(out)
// }

// // parse "HH:MM" into time.Time UTC using today's date
// func ParseHHMMToTimeUTC(hm string) (time.Time, error) {
// 	parts := SplitAndTrimAny(hm, []string{":", "."})
// 	if len(parts) < 2 {
// 		return time.Time{}, fmt.Errorf("invalid time")
// 	}
// 	h, err1 := strconv.Atoi(parts[0])
// 	m, err2 := strconv.Atoi(parts[1])
// 	if err1 != nil || err2 != nil {
// 		return time.Time{}, fmt.Errorf("invalid time ints")
// 	}
// 	now := time.Now().UTC()
// 	t := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, time.UTC)
// 	return t, nil
// }

// // split on any of the provided separators and trim
// func SplitAndTrimAny(s string, seps []string) []string {
// 	for _, sp := range seps {
// 		if strings.Contains(s, sp) {
// 			parts := strings.Split(s, sp)
// 			out := []string{}
// 			for _, p := range parts {
// 				t := strings.TrimSpace(p)
// 				if t != "" {
// 					out = append(out, t)
// 				}
// 			}
// 			return out
// 		}
// 	}
// 	// fallback to single value
// 	if strings.TrimSpace(s) == "" {
// 		return []string{}
// 	}
// 	return []string{strings.TrimSpace(s)}
// }
/**
package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"

	"github.com/parshwanath-p2493/GreenCartt/config"
	"github.com/parshwanath-p2493/GreenCartt/models"
)

func SeedFromExcel(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Drop collections first
	_ = config.DB.Collection("drivers").Drop(ctx)
	_ = config.DB.Collection("routes").Drop(ctx)
	_ = config.DB.Collection("orders").Drop(ctx)

	// 1) Load and seed routes.xlsx
	xlRoutes, err := excelize.OpenFile("backend/data/routes.csv")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "cannot open routes.xlsx", "details": err.Error()})
	}
	rows, err := xlRoutes.GetRows("routes") // or correct sheet name in routes.xlsx
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "cannot read rows from routes.xlsx", "details": err.Error()})
	}
	var routes []interface{}
	for i, r := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(r) < 4 {
			continue
		}
		routeID := strings.TrimSpace(r[0])
		distance := parseFloatOrZero(r[1])
		traffic := strings.TrimSpace(r[2])
		baseMin := parseFloatOrZero(r[3])
		route := models.Route{
			RouteID:      routeID,
			DistanceKm:   distance,
			TrafficLevel: traffic,
			BaseTimeMin:  baseMin,
		}
		routes = append(routes, route)
	}
	if len(routes) > 0 {
		if _, err := config.DB.Collection("routes").InsertMany(ctx, routes); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "routes insert failed", "details": err.Error()})
		}
	}

	// 2) Load and seed drivers.csv
	xlDrivers, err := excelize.OpenFile("backend/data/drivers.csv")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "cannot open drivers.csv", "details": err.Error()})
	}
	rows, err = xlDrivers.GetRows("drivers") // or correct sheet name in drivers.csv
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "cannot read rows from drivers.csv", "details": err.Error()})
	}
	var drivers []interface{}
	for i, r := range rows {
		if i == 0 {
			continue
		}
		if len(r) < 3 {
			continue
		}
		name := strings.TrimSpace(r[0])
		shiftHrs := parseFloatOrZero(r[1])
		rawPast := strings.TrimSpace(r[2]) // e.g., "6|8|7|7|7|6|10"
		parts := splitAndTrimAny(rawPast, []string{"|", ","})
		past := []int{}
		for _, p := range parts {
			if p == "" {
				continue
			}
			v, err := strconv.Atoi(p)
			if err == nil {
				past = append(past, v)
			}
		}
		id := makeDriverID(name)
		d := models.Driver{
			ID:              id,
			Name:            name,
			CurrentShiftHrs: shiftHrs,
			Past7DayHours:   past,
		}
		drivers = append(drivers, d)
	}
	if len(drivers) > 0 {
		if _, err := config.DB.Collection("drivers").InsertMany(ctx, drivers); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "drivers insert failed", "details": err.Error()})
		}
	}

	// 3) Load and seed orders.csv
	xlOrders, err := excelize.OpenFile("backend/data/orders.csv")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "cannot open orders.csv", "details": err.Error()})
	}
	rows, err = xlOrders.GetRows("orders") // or correct sheet name in orders.csv
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "cannot read rows from orders.csv", "details": err.Error()})
	}
	var orders []interface{}
	for i, r := range rows {
		if i == 0 {
			continue
		}
		if len(r) < 4 {
			continue
		}
		orderID := strings.TrimSpace(r[0])
		value := parseFloatOrZero(r[1])
		routeID := strings.TrimSpace(r[2])
		dtime := strings.TrimSpace(r[3]) // "02:07"
		ts, perr := parseHHMMToTimeUTC(dtime)
		if perr != nil {
			ts = time.Now().UTC()
		}
		o := models.Order{
			OrderID:        orderID,
			ValueRs:        value,
			AssignedRoute:  routeID,
			DeliveryTimeTs: ts,
		}
		orders = append(orders, o)
	}
	if len(orders) > 0 {
		if _, err := config.DB.Collection("orders").InsertMany(ctx, orders); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "orders insert failed", "details": err.Error()})
		}
	}

	return c.JSON(fiber.Map{"status": "ok", "message": "seed completed"})
}

func parseFloatOrZero(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func splitAndTrimAny(s string, seps []string) []string {
	for _, sp := range seps {
		if strings.Contains(s, sp) {
			parts := strings.Split(s, sp)
			out := []string{}
			for _, p := range parts {
				t := strings.TrimSpace(p)
				if t != "" {
					out = append(out, t)
				}
			}
			return out
		}
	}
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	return []string{strings.TrimSpace(s)}
}

func makeDriverID(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	out := []rune{}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			out = append(out, r)
		}
	}
	return string(out)
}

// parse "HH:MM" into time.Time UTC using today's date
func parseHHMMToTimeUTC(hm string) (time.Time, error) {
	parts := splitAndTrimAny(hm, []string{":", "."})
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid time")
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return time.Time{}, fmt.Errorf("invalid time ints")
	}
	now := time.Now().UTC()
	t := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, time.UTC)
	return t, nil
}
**/

package handlers

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/parshwanath-p2493/GreenCartt/config"
	"github.com/parshwanath-p2493/GreenCartt/models"
)

func SeedFromExcel(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Drop collections first
	_ = config.DB.Collection("drivers").Drop(ctx)
	_ = config.DB.Collection("routes").Drop(ctx)
	_ = config.DB.Collection("orders").Drop(ctx)

	// Helper function to read CSV
	readCSV := func(path string) ([][]string, error) {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		r := csv.NewReader(f)
		return r.ReadAll()
	}

	// 1) Load and seed routes.csv
	rows, err := readCSV("data/routes.csv")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "cannot open routes.csv", "details": err.Error()})
	}
	var routes []interface{}
	for i, r := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(r) < 4 {
			continue
		}
		routeID := strings.TrimSpace(r[0])
		distance := parseFloatOrZero(r[1])
		traffic := strings.TrimSpace(r[2])
		baseMin := parseFloatOrZero(r[3])
		route := models.Route{
			RouteID:      routeID,
			DistanceKm:   distance,
			TrafficLevel: traffic,
			BaseTimeMin:  baseMin,
		}
		routes = append(routes, route)
	}
	if len(routes) > 0 {
		if _, err := config.DB.Collection("routes").InsertMany(ctx, routes); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "routes insert failed", "details": err.Error()})
		}
	}

	// 2) Load and seed drivers.csv
	rows, err = readCSV("data/drivers.csv")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "cannot open drivers.csv", "details": err.Error()})
	}
	var drivers []interface{}
	for i, r := range rows {
		if i == 0 {
			continue
		}
		if len(r) < 3 {
			continue
		}
		name := strings.TrimSpace(r[0])
		shiftHrs := parseFloatOrZero(r[1])
		rawPast := strings.TrimSpace(r[2]) // e.g., "6|8|7|7|7|6|10"
		parts := splitAndTrimAny(rawPast, []string{"|", ","})
		past := []int{}
		for _, p := range parts {
			if p == "" {
				continue
			}
			v, err := strconv.Atoi(p)
			if err == nil {
				past = append(past, v)
			}
		}
		id := makeDriverID(name)
		d := models.Driver{
			ID:              id,
			Name:            name,
			CurrentShiftHrs: shiftHrs,
			Past7DayHours:   past,
		}
		drivers = append(drivers, d)
	}
	if len(drivers) > 0 {
		if _, err := config.DB.Collection("drivers").InsertMany(ctx, drivers); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "drivers insert failed", "details": err.Error()})
		}
	}

	// 3) Load and seed orders.csv
	rows, err = readCSV("data/orders.csv")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "cannot open orders.csv", "details": err.Error()})
	}
	var orders []interface{}
	for i, r := range rows {
		if i == 0 {
			continue
		}
		if len(r) < 4 {
			continue
		}
		orderID := strings.TrimSpace(r[0])
		value := parseFloatOrZero(r[1])
		routeID := strings.TrimSpace(r[2])
		dtime := strings.TrimSpace(r[3]) // "02:07"
		ts, perr := parseHHMMToTimeUTC(dtime)
		if perr != nil {
			ts = time.Now().UTC()
		}
		o := models.Order{
			OrderID:        orderID,
			ValueRs:        value,
			AssignedRoute:  routeID,
			DeliveryTimeTs: ts,
		}
		orders = append(orders, o)
	}
	if len(orders) > 0 {
		if _, err := config.DB.Collection("orders").InsertMany(ctx, orders); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "orders insert failed", "details": err.Error()})
		}
	}

	return c.JSON(fiber.Map{"status": "ok", "message": "seed completed"})
}

// Keep your helper functions parseFloatOrZero, splitAndTrimAny, makeDriverID, parseHHMMToTimeUTC unchanged

func parseFloatOrZero(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func splitAndTrimAny(s string, seps []string) []string {
	for _, sp := range seps {
		if strings.Contains(s, sp) {
			parts := strings.Split(s, sp)
			out := []string{}
			for _, p := range parts {
				t := strings.TrimSpace(p)
				if t != "" {
					out = append(out, t)
				}
			}
			return out
		}
	}
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	return []string{strings.TrimSpace(s)}
}

func makeDriverID(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	out := []rune{}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			out = append(out, r)
		}
	}
	return string(out)
}

// parse "HH:MM" into time.Time UTC using today's date
func parseHHMMToTimeUTC(hm string) (time.Time, error) {
	parts := splitAndTrimAny(hm, []string{":", "."})
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid time")
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return time.Time{}, fmt.Errorf("invalid time ints")
	}
	now := time.Now().UTC()
	t := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, time.UTC)
	return t, nil
}
