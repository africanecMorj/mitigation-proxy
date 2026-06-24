package config

type Matcher interface {
    Match() balancers.Balancer
}