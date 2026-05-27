package role

type Name string

var System = struct {
	Admin    Name
	Employee Name
}{
	Admin:    "admin",
	Employee: "employee",
}
