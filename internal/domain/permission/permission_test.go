package permission

import "testing"

func TestPermissionValues_NonEmpty(t *testing.T) {
	all := []Permission{
		Employee.Create, Employee.Delete, Employee.Update, Employee.View,
		Portal.AdminAccess, Portal.EmployeeAccess,
		Handbook.Assign,
		Handbook.Self.Update, Handbook.Self.View,
		Handbook.Step.Create, Handbook.Step.Delete, Handbook.Step.Update, Handbook.Step.View,
		Handbook.Template.Create, Handbook.Template.Publish, Handbook.Template.Update, Handbook.Template.View,
		LateArrival.Create, LateArrival.CreateAll, LateArrival.View, LateArrival.ViewAll,
		Leave.Balance.Adjust, Leave.Balance.View, Leave.Balance.ViewAll,
		Leave.Request.Create, Leave.Request.Decide, Leave.Request.Update, Leave.Request.UpdateAll, Leave.Request.View, Leave.Request.ViewAll,
		Location.Create, Location.Delete, Location.Update, Location.View,
		Payout.Request.Create, Payout.Request.Decide, Payout.Request.MarkPaid, Payout.Request.View, Payout.Request.ViewAll,
		Expense.Request.Create, Expense.Request.Decide, Expense.Request.MarkReimbursed, Expense.Request.Update, Expense.Request.View, Expense.Request.ViewAll,
		PayPeriod.Close, PayPeriod.MarkPaid, PayPeriod.MonthSummaryView, PayPeriod.ViewAll,
		Role.View,
		Performance.Assessment.Create, Performance.Assessment.Delete, Performance.Assessment.View, Performance.Assessment.ViewAll,
		Performance.Stats,
		Performance.Upcoming.Invite,
		Performance.WorkAssignment.Decide, Performance.WorkAssignment.View, Performance.WorkAssignment.ViewAll,
		Training.Catalog.Create, Training.Catalog.View,
		Training.Assign, Training.AssignmentsView,
		Schedule.Create, Schedule.Delete, Schedule.Update, Schedule.View,
		Settings.View, Settings.Update,
		ScheduleSwap.Approve, ScheduleSwap.Request, ScheduleSwap.Respond, ScheduleSwap.View,
		Shift.Create, Shift.Delete, Shift.Update, Shift.View,
	}

	seen := make(map[Permission]struct{}, len(all))
	for _, p := range all {
		if p == "" {
			t.Error("empty permission value found")
		}
		if _, ok := seen[p]; ok {
			t.Errorf("duplicate permission value: %q", p)
		}
		seen[p] = struct{}{}
	}
}
