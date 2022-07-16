package route

type AdministrativeDistance uint8

const (
	ADConnected AdministrativeDistance = 0
	ADStatic    AdministrativeDistance = 1
	ADEBGP      AdministrativeDistance = 20
	ADOSPF      AdministrativeDistance = 110
	ADRIP       AdministrativeDistance = 120
	ADIBGP      AdministrativeDistance = 200
)
