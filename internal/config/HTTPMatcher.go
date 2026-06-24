package config

type HTTPMatcher struct {
    ExactHosts      map[string]balancers.Balancer
    WildcardHosts   []WildcardRule
    DefaultBalancer balancers.Balancer
}

func (p *Picker) Match(
    host string,
) balancers.Balancer {

   

    return p.DefaultBalancer
}