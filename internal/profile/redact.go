package profile

const redacted = "********"

// Redacted returns a copy of p with every password replaced by a placeholder,
// for printing a profile without leaking its secrets. p itself is unchanged.
func (p *Profile) Redacted() *Profile {
	c := *p
	mask := func(s *string) *string {
		if s == nil {
			return nil
		}
		m := redacted
		return &m
	}
	if p.Accounts != nil {
		c.Accounts = make([]UserAccount, len(p.Accounts))
		for i, a := range p.Accounts {
			a.Password = mask(a.Password)
			c.Accounts[i] = a
		}
	}
	c.FirstLogon.BuiltinAdministratorPassword = mask(p.FirstLogon.BuiltinAdministratorPassword)
	if p.Wifi != nil {
		w := *p.Wifi
		w.Password = mask(w.Password)
		c.Wifi = &w
	}
	return &c
}
