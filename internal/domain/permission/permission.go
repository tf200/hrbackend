package permission

type Permission string

var Employee = struct {
	Create Permission
	Delete Permission
	Update Permission
	View   Permission
}{
	Create: "EMPLOYEE.CREATE",
	Delete: "EMPLOYEE.DELETE",
	Update: "EMPLOYEE.UPDATE",
	View:   "EMPLOYEE.VIEW",
}

var Portal = struct {
	AdminAccess    Permission
	EmployeeAccess Permission
}{
	AdminAccess:    "PORTAL.ADMIN.ACCESS",
	EmployeeAccess: "PORTAL.EMPLOYEE.ACCESS",
}

var Handbook = struct {
	Assign Permission
	Self   struct {
		Update Permission
		View   Permission
	}
	Step struct {
		Create Permission
		Delete Permission
		Update Permission
		View   Permission
	}
	Template struct {
		Create  Permission
		Publish Permission
		Update  Permission
		View    Permission
	}
}{
	Assign: "HANDBOOK.ASSIGN",
	Self: struct {
		Update Permission
		View   Permission
	}{
		Update: "HANDBOOK.SELF.UPDATE",
		View:   "HANDBOOK.SELF.VIEW",
	},
	Step: struct {
		Create Permission
		Delete Permission
		Update Permission
		View   Permission
	}{
		Create: "HANDBOOK.STEP.CREATE",
		Delete: "HANDBOOK.STEP.DELETE",
		Update: "HANDBOOK.STEP.UPDATE",
		View:   "HANDBOOK.STEP.VIEW",
	},
	Template: struct {
		Create  Permission
		Publish Permission
		Update  Permission
		View    Permission
	}{
		Create:  "HANDBOOK.TEMPLATE.CREATE",
		Publish: "HANDBOOK.TEMPLATE.PUBLISH",
		Update:  "HANDBOOK.TEMPLATE.UPDATE",
		View:    "HANDBOOK.TEMPLATE.VIEW",
	},
}

var LateArrival = struct {
	Create   Permission
	CreateAll Permission
	View     Permission
	ViewAll  Permission
}{
	Create:    "LATE_ARRIVAL.CREATE",
	CreateAll: "LATE_ARRIVAL.CREATE_ALL",
	View:      "LATE_ARRIVAL.VIEW",
	ViewAll:   "LATE_ARRIVAL.VIEW_ALL",
}

var Leave = struct {
	Balance struct {
		Adjust  Permission
		View    Permission
		ViewAll Permission
	}
	Request struct {
		Create    Permission
		Decide    Permission
		Update    Permission
		UpdateAll Permission
		View      Permission
		ViewAll   Permission
	}
}{
	Balance: struct {
		Adjust  Permission
		View    Permission
		ViewAll Permission
	}{
		Adjust:  "LEAVE.BALANCE.ADJUST",
		View:    "LEAVE.BALANCE.VIEW",
		ViewAll: "LEAVE.BALANCE.VIEW_ALL",
	},
	Request: struct {
		Create    Permission
		Decide    Permission
		Update    Permission
		UpdateAll Permission
		View      Permission
		ViewAll   Permission
	}{
		Create:    "LEAVE.REQUEST.CREATE",
		Decide:    "LEAVE.REQUEST.DECIDE",
		Update:    "LEAVE.REQUEST.UPDATE",
		UpdateAll: "LEAVE.REQUEST.UPDATE_ALL",
		View:      "LEAVE.REQUEST.VIEW",
		ViewAll:   "LEAVE.REQUEST.VIEW_ALL",
	},
}

var Location = struct {
	Create Permission
	Delete Permission
	Update Permission
	View   Permission
}{
	Create: "LOCATION.CREATE",
	Delete: "LOCATION.DELETE",
	Update: "LOCATION.UPDATE",
	View:   "LOCATION.VIEW",
}

var Payout = struct {
	Request struct {
		Create   Permission
		Decide   Permission
		MarkPaid Permission
		View     Permission
		ViewAll  Permission
	}
}{
	Request: struct {
		Create   Permission
		Decide   Permission
		MarkPaid Permission
		View     Permission
		ViewAll  Permission
	}{
		Create:   "PAYOUT.REQUEST.CREATE",
		Decide:   "PAYOUT.REQUEST.DECIDE",
		MarkPaid: "PAYOUT.REQUEST.MARK_PAID",
		View:     "PAYOUT.REQUEST.VIEW",
		ViewAll:  "PAYOUT.REQUEST.VIEW_ALL",
	},
}

var Expense = struct {
	Request struct {
		Create         Permission
		Decide         Permission
		MarkReimbursed Permission
		Update         Permission
		View           Permission
		ViewAll        Permission
	}
}{
	Request: struct {
		Create         Permission
		Decide         Permission
		MarkReimbursed Permission
		Update         Permission
		View           Permission
		ViewAll        Permission
	}{
		Create:         "EXPENSE.REQUEST.CREATE",
		Decide:         "EXPENSE.REQUEST.DECIDE",
		MarkReimbursed: "EXPENSE.REQUEST.MARK_REIMBURSED",
		Update:         "EXPENSE.REQUEST.UPDATE",
		View:           "EXPENSE.REQUEST.VIEW",
		ViewAll:        "EXPENSE.REQUEST.VIEW_ALL",
	},
}

var PayPeriod = struct {
	Close             Permission
	MarkPaid          Permission
	MonthSummaryView  Permission
	ViewAll           Permission
}{
	Close:            "PAY_PERIOD.CLOSE",
	MarkPaid:         "PAY_PERIOD.MARK_PAID",
	MonthSummaryView: "PAY_PERIOD.MONTH_SUMMARY_VIEW",
	ViewAll:          "PAY_PERIOD.VIEW_ALL",
}

var Role = struct {
	View Permission
}{
	View: "ROLE.VIEW",
}

var Performance = struct {
	Assessment struct {
		Create Permission
		Delete Permission
		View   Permission
		ViewAll Permission
	}
	Stats     Permission
	Upcoming  struct {
		Invite Permission
	}
	WorkAssignment struct {
		Decide  Permission
		View    Permission
		ViewAll Permission
	}
}{
	Assessment: struct {
		Create  Permission
		Delete  Permission
		View    Permission
		ViewAll Permission
	}{
		Create:  "PERFORMANCE.ASSESSMENT.CREATE",
		Delete:  "PERFORMANCE.ASSESSMENT.DELETE",
		View:    "PERFORMANCE.ASSESSMENT.VIEW",
		ViewAll: "PERFORMANCE.ASSESSMENT.VIEW_ALL",
	},
	Stats: "PERFORMANCE.STATS.VIEW",
	Upcoming: struct {
		Invite Permission
	}{
		Invite: "PERFORMANCE.UPCOMING.INVITE",
	},
	WorkAssignment: struct {
		Decide  Permission
		View    Permission
		ViewAll Permission
	}{
		Decide:  "PERFORMANCE.WORK_ASSIGNMENT.DECIDE",
		View:    "PERFORMANCE.WORK_ASSIGNMENT.VIEW",
		ViewAll: "PERFORMANCE.WORK_ASSIGNMENT.VIEW_ALL",
	},
}

var Training = struct {
	Catalog struct {
		Create Permission
		View   Permission
	}
	Assign         Permission
	AssignmentsView Permission
}{
	Catalog: struct {
		Create Permission
		View   Permission
	}{
		Create: "TRAINING.CATALOG.CREATE",
		View:   "TRAINING.CATALOG.VIEW",
	},
	Assign:          "TRAINING.ASSIGN",
	AssignmentsView: "TRAINING.ASSIGNMENTS.VIEW",
}

var Schedule = struct {
	Create Permission
	Delete Permission
	Update Permission
	View   Permission
}{
	Create: "SCHEDULE.CREATE",
	Delete: "SCHEDULE.DELETE",
	Update: "SCHEDULE.UPDATE",
	View:   "SCHEDULE.VIEW",
}

var Settings = struct {
	View   Permission
	Update Permission
}{
	View:   "SETTINGS.VIEW",
	Update: "SETTINGS.UPDATE",
}

var ScheduleSwap = struct {
	Approve Permission
	Request Permission
	Respond Permission
	View    Permission
}{
	Approve: "SCHEDULE_SWAP.APPROVE",
	Request: "SCHEDULE_SWAP.REQUEST",
	Respond: "SCHEDULE_SWAP.RESPOND",
	View:    "SCHEDULE_SWAP.VIEW",
}

var Shift = struct {
	Create Permission
	Delete Permission
	Update Permission
	View   Permission
}{
	Create: "SHIFT.CREATE",
	Delete: "SHIFT.DELETE",
	Update: "SHIFT.UPDATE",
	View:   "SHIFT.VIEW",
}

var Overtime = struct {
	Create    Permission
	CreateAll Permission
	Decide    Permission
	Update    Permission
	UpdateAll Permission
	View      Permission
	ViewAll   Permission
}{
	Create:    "OVERTIME.CREATE",
	CreateAll: "OVERTIME.CREATE_ALL",
	Decide:    "OVERTIME.DECIDE",
	Update:    "OVERTIME.UPDATE",
	UpdateAll: "OVERTIME.UPDATE_ALL",
	View:      "OVERTIME.VIEW",
	ViewAll:   "OVERTIME.VIEW_ALL",
}

var TimeEntry = struct {
	Create    Permission
	CreateAll Permission
	Decide    Permission
	Update    Permission
	UpdateAll Permission
	View      Permission
	ViewAll   Permission
}{
	Create:    "TIME_ENTRY.CREATE",
	CreateAll: "TIME_ENTRY.CREATE_ALL",
	Decide:    "TIME_ENTRY.DECIDE",
	Update:    "TIME_ENTRY.UPDATE",
	UpdateAll: "TIME_ENTRY.UPDATE_ALL",
	View:      "TIME_ENTRY.VIEW",
	ViewAll:   "TIME_ENTRY.VIEW_ALL",
}
