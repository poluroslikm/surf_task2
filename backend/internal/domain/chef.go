package domain

// Chef is a read-only catalog entry (data-model.md → Chef). Regular and guest chefs are the
// same resource, no separate status (02-domain.md).
type Chef struct {
	ID   string
	Name string
}
