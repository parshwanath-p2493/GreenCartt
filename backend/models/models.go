package models

import "time"

type Driver struct {
	ID              string  `bson:"id" json:"id"` // we use custom id string for easy seeding
	Name            string  `bson:"name" json:"name"`
	CurrentShiftHrs float64 `bson:"current_shift_hours" json:"current_shift_hours"`
	Past7DayHours   []int   `bson:"past_7_day_hours" json:"past_7_day_hours"`
}

type Route struct {
	RouteID      string  `bson:"route_id" json:"route_id"`
	DistanceKm   float64 `bson:"distance_km" json:"distance_km"`
	TrafficLevel string  `bson:"traffic_level" json:"traffic_level"` // "Low"|"Medium"|"High"
	BaseTimeMin  float64 `bson:"base_time_minutes" json:"base_time_minutes"`
}

type Order struct {
	OrderID        string    `bson:"order_id" json:"order_id"`
	ValueRs        float64   `bson:"value_rs" json:"value_rs"`
	AssignedRoute  string    `bson:"assigned_route_id" json:"assigned_route_id"`
	DeliveryTimeTs time.Time `bson:"delivery_timestamp" json:"delivery_timestamp"`
}

type SimulationInput struct {
	AvailableDrivers  int    `json:"available_drivers"`
	RouteStartTime    string `json:"route_start_time"` // HH:MM
	MaxHoursPerDriver int    `json:"max_hours_per_driver"`
}

type OrderResult struct {
	OrderID         string  `bson:"order_id" json:"order_id"`
	AssignedDriver  string  `bson:"assigned_driver" json:"assigned_driver"`
	DeliveredOnTime bool    `bson:"delivered_on_time" json:"delivered_on_time"`
	PenaltyRs       float64 `bson:"penalty_rs" json:"penalty_rs"`
	BonusRs         float64 `bson:"bonus_rs" json:"bonus_rs"`
	FuelCostRs      float64 `bson:"fuel_cost_rs" json:"fuel_cost_rs"`
	ProfitRs        float64 `bson:"profit_rs" json:"profit_rs"`
	ExpectedMinutes float64 `bson:"expected_minutes" json:"expected_minutes"`
}

type SimulationResult struct {
	ID                string             `bson:"_id,omitempty" json:"simulation_id"`
	Timestamp         time.Time          `bson:"timestamp" json:"timestamp"`
	Inputs            SimulationInput    `bson:"inputs" json:"inputs"`
	TotalProfitRs     float64            `bson:"total_profit_rs" json:"total_profit_rs"`
	EfficiencyScore   float64            `bson:"efficiency_score" json:"efficiency_score"`
	TotalOrders       int                `bson:"total_orders" json:"total_orders"`
	OnTime            int                `bson:"on_time" json:"on_time"`
	Late              int                `bson:"late" json:"late"`
	FuelCostBreakdown map[string]float64 `bson:"fuel_cost_breakdown" json:"fuel_cost_breakdown"`
	OrderResults      []OrderResult      `bson:"order_results" json:"order_results"`
}
