package profile

import "testing"

func TestRedactedMasksPasswordsWithoutTouchingTheOriginal(t *testing.T) {
	pw, admin, wifi := "p1", "p2", "p3"
	p := Default("demo")
	p.Accounts = []UserAccount{{Name: "u", Password: &pw, Group: GroupUsers}, {Name: "v", Group: GroupUsers}}
	p.FirstLogon.BuiltinAdministratorPassword = &admin
	p.Wifi = &WifiSettings{SSID: "n", Authentication: WifiWPA2Personal, Password: &wifi}

	r := p.Redacted()
	if *r.Accounts[0].Password != redacted || r.Accounts[1].Password != nil {
		t.Fatalf("accounts not masked correctly: %+v", r.Accounts)
	}
	if *r.FirstLogon.BuiltinAdministratorPassword != redacted || *r.Wifi.Password != redacted {
		t.Fatal("administrator or Wi-Fi password not masked")
	}
	if *p.Accounts[0].Password != "p1" || *p.FirstLogon.BuiltinAdministratorPassword != "p2" || *p.Wifi.Password != "p3" {
		t.Fatal("Redacted modified the original profile")
	}
}
