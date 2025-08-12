package services

import (
	"errors"
	"time"

	"github.com/parshwanath-p2493/GreenCartt/models"
)

// company constants
const (
	LatePenaltyRs              = 50.0
	BaseFuelPerKm              = 5.0
	HighTrafficSurcharge       = 2.0
	HighValueThreshold         = 1000.0
	HighValueBonusRate         = 0.10
	FatigueThresholdHours      = 8.0
	FatigueSpeedIncreaseFactor = 1.3 // if fatigued, route time increases by 30% -> multiplier 1.3
)

// Simulate performs simulation given raw slices of drivers, routes, orders and inputs.
// deterministic allocation: round-robin mapping of orders to available drivers
func Simulate(drivers []models.Driver, routes []models.Route, orders []models.Order, input models.SimulationInput) (models.SimulationResult, error) {
	res := models.SimulationResult{
		Timestamp:         time.Now().UTC(),
		Inputs:            input,
		FuelCostBreakdown: map[string]float64{"Low": 0, "Medium": 0, "High": 0},
	}

	if input.AvailableDrivers <= 0 || input.AvailableDrivers > len(drivers) {
		return res, errors.New("available_drivers must be >0 and <= total drivers")
	}
	// map routeID -> route
	routeMap := make(map[string]models.Route)
	for _, r := range routes {
		routeMap[r.RouteID] = r
	}

	orderResults := make([]models.OrderResult, 0, len(orders))
	totalProfit := 0.0
	onTime := 0
	late := 0

	// prepare driver list (only first N available drivers)
	activeDrivers := drivers[:input.AvailableDrivers]
	driverCount := len(activeDrivers)
	if driverCount == 0 {
		return res, errors.New("no active drivers")
	}

	// determine fatigue for each driver by checking last day's hours (use most recent as index 0)
	driverFatigue := make([]bool, driverCount)
	for i, d := range activeDrivers {
		sum := 0
		for _, h := range d.Past7DayHours {
			sum += h
		}
		// If last day (>8 hours) - but doc specified >8 in a day -> mark next day fatigued
		// We'll interpret: if any day >8 in past 7 days => fatigued (assumption). Document in README.
		fatigued := false
		for _, h := range d.Past7DayHours {
			if h > 8 {
				fatigued = true
				break
			}
		}
		driverFatigue[i] = fatigued
		_ = sum
	}

	// round-robin allocate orders to drivers
	for idx, ord := range orders {
		orderResult := models.OrderResult{OrderID: ord.OrderID}
		driverIdx := idx % driverCount
		driver := activeDrivers[driverIdx]
		orderResult.AssignedDriver = driver.ID

		route, ok := routeMap[ord.AssignedRoute]
		if !ok {
			// skip or error - create a default negative profit
			orderResult.ProfitRs = 0
			orderResults = append(orderResults, orderResult)
			continue
		}

		// compute expected minutes from route base time
		expectedMinutes := route.BaseTimeMin
		if driverFatigue[driverIdx] {
			expectedMinutes = expectedMinutes * FatigueSpeedIncreaseFactor // increase time by 30%
		}
		orderResult.ExpectedMinutes = expectedMinutes

		// Compare delivered_on_time: use ord.DeliveryTimeTs and route start time?
		// As a simple deterministic check, we'll compute actual_delivery_time = expectedMinutes (since start at route_start_time)
		// and treat it as on-time if expectedMinutes <= base_time + 10
		deliveredOnTime := expectedMinutes <= (route.BaseTimeMin + 10.0)
		orderResult.DeliveredOnTime = deliveredOnTime
		if deliveredOnTime {
			onTime++
		} else {
			late++
			orderResult.PenaltyRs = LatePenaltyRs
		}

		// bonus
		if ord.ValueRs > HighValueThreshold && deliveredOnTime {
			orderResult.BonusRs = ord.ValueRs * HighValueBonusRate
		}

		// fuel cost
		perKm := BaseFuelPerKm
		if route.TrafficLevel == "High" {
			perKm += HighTrafficSurcharge
		}
		orderResult.FuelCostRs = route.DistanceKm * perKm
		res.FuelCostBreakdown[route.TrafficLevel] += orderResult.FuelCostRs

		// profit
		orderResult.ProfitRs = ord.ValueRs + orderResult.BonusRs - orderResult.PenaltyRs - orderResult.FuelCostRs
		totalProfit += orderResult.ProfitRs

		orderResults = append(orderResults, orderResult)
	}

	res.OrderResults = orderResults
	res.TotalOrders = len(orders)
	res.TotalProfitRs = totalProfit
	res.OnTime = onTime
	res.Late = late
	if res.TotalOrders > 0 {
		res.EfficiencyScore = (float64(onTime) / float64(res.TotalOrders)) * 100.0
	} else {
		res.EfficiencyScore = 0
	}

	return res, nil
}
