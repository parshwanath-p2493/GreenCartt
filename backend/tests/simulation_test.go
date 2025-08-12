package tests

import (
	"testing"
	"time"

	"github.com/parshwanath-p2493/GreenCartt/models"
	"github.com/parshwanath-p2493/GreenCartt/services"
)

func sampleData() ([]models.Driver, []models.Route, []models.Order) {
	drivers := []models.Driver{
		{ID: "d1", Name: "A", CurrentShiftHrs: 0, Past7DayHours: []int{8, 7, 6, 6, 6, 6, 6}},
		{ID: "d2", Name: "B", CurrentShiftHrs: 0, Past7DayHours: []int{9, 9, 9, 9, 9, 9, 9}}, // fatigued
	}
	routes := []models.Route{
		{RouteID: "r1", DistanceKm: 10, TrafficLevel: "High", BaseTimeMin: 30},
		{RouteID: "r2", DistanceKm: 5, TrafficLevel: "Low", BaseTimeMin: 15},
	}
	orders := []models.Order{
		{OrderID: "o1", ValueRs: 1200, AssignedRoute: "r1", DeliveryTimeTs: time.Now()},
		{OrderID: "o2", ValueRs: 300, AssignedRoute: "r2", DeliveryTimeTs: time.Now()},
	}
	return drivers, routes, orders
}

func TestSimulateBasic(t *testing.T) {
	drivers, routes, orders := sampleData()
	input := models.SimulationInput{AvailableDrivers: 2, RouteStartTime: "08:30", MaxHoursPerDriver: 9}
	res, err := services.Simulate(drivers, routes, orders, input)
	if err != nil {
		t.Fatalf("simulate failed: %v", err)
	}
	if res.TotalOrders != 2 {
		t.Fatalf("expected 2 orders, got %d", res.TotalOrders)
	}
	// check efficiency between 0 and 100
	if res.EfficiencyScore < 0 || res.EfficiencyScore > 100 {
		t.Fatalf("eff score out of range: %f", res.EfficiencyScore)
	}
}

func TestFatigueEffect(t *testing.T) {
	drivers, routes, orders := sampleData()
	input := models.SimulationInput{AvailableDrivers: 2, RouteStartTime: "08:30", MaxHoursPerDriver: 9}
	res, _ := services.Simulate(drivers, routes, orders, input)
	// driver 2 is fatigued (all 9s), so second assigned order expectedMinutes should be > base
	if len(res.OrderResults) < 2 {
		t.Fatalf("not enough results")
	}
	_ = res.OrderResults[0]
	second := res.OrderResults[1]
	if second.ExpectedMinutes <= routes[1].BaseTimeMin {
		t.Fatalf("fatigue not applied: %f <= %f", second.ExpectedMinutes, routes[1].BaseTimeMin)
	}
}

func TestFuelCostCalc(t *testing.T) {
	_, routes, orders := sampleData()
	input := models.SimulationInput{AvailableDrivers: 1, RouteStartTime: "08:30", MaxHoursPerDriver: 9}
	drivers := []models.Driver{{ID: "d1", Name: "A", Past7DayHours: []int{6, 6, 6, 6, 6, 6, 6}}}
	res, _ := services.Simulate(drivers, routes, orders, input)
	// check fuel cost recorded in breakdown
	if res.FuelCostBreakdown["High"] <= 0 {
		t.Fatalf("expected high traffic fuel cost >0")
	}
}

func TestBonusApplied(t *testing.T) {
	drivers, routes, orders := sampleData()
	input := models.SimulationInput{AvailableDrivers: 2, RouteStartTime: "08:30", MaxHoursPerDriver: 9}
	res, _ := services.Simulate(drivers, routes, orders, input)
	// find order with value>1000 and on-time => bonus >0
	found := false
	for _, or := range res.OrderResults {
		if or.OrderID == "o1" {
			if or.BonusRs <= 0 && or.DeliveredOnTime {
				t.Fatalf("expected bonus for o1")
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("o1 not found")
	}
}
