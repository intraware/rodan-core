package models

import "gorm.io/gorm"

type Challenge struct {
	gorm.Model
	ID           int      `json:"id" gorm:"unique;index"`
	Name         string   `json:"name"`
	Desc         string   `json:"desc"`
	Category     int8     `json:"category"`
	IsStatic     bool     `json:"is_static"`
	Flag         string   `json:"flag,omitempty"`          // Only for static challenges
	Ports        []int    `json:"ports,omitempty"`         // Only for static challenges
	Links        []string `json:"links,omitempty"`         // Only for static challenges
	DockerImage  string   `json:"docker_image,omitempty"`  // Docker image for dynamic challenges
	ExposedPorts []string `json:"exposed_ports,omitempty"` // Container ports for dynamic challenges
	PointsMin    int      `json:"points_min"`
	PointsMax    int      `json:"points_max"`
	Difficulty   int8     `json:"difficulty"`
}
