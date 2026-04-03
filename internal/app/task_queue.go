package app

import (
	"context"

	"hrbackend/internal/domain"
	pkgasynq "hrbackend/pkg/asynq"

	hibikenasynq "github.com/hibiken/asynq"
)

type taskQueueAdapter struct {
	client *pkgasynq.AsynqClient
}

func (a *taskQueueAdapter) EnqueueEmailDelivery(
	ctx context.Context,
	payload domain.EmailDeliveryTaskPayload,
	opts *domain.TaskEnqueueOptions,
) error {
	return a.client.EnqueueEmailDelivery(
		toEmailDeliveryPayload(payload),
		ctx,
		toAsynqOptions(opts)...)
}

func (a *taskQueueAdapter) EnqueueIncident(
	ctx context.Context,
	payload domain.IncidentTaskPayload,
	opts *domain.TaskEnqueueOptions,
) error {
	return a.client.EnqueueIncident(toIncidentPayload(payload), ctx, toAsynqOptions(opts)...)
}

func (a *taskQueueAdapter) EnqueueIncidentConfirmedEmail(
	ctx context.Context,
	payload domain.IncidentConfirmedEmailTaskPayload,
	opts *domain.TaskEnqueueOptions,
) error {
	return a.client.EnqueueIncidentConfirmedEmail(ctx, pkgasynq.IncidentConfirmedEmailPayload{
		IncidentID: payload.IncidentID,
	}, toAsynqOptions(opts)...)
}

func (a *taskQueueAdapter) EnqueueNotificationTask(
	ctx context.Context,
	payload domain.NotificationTaskPayload,
	opts *domain.TaskEnqueueOptions,
) error {
	return a.client.EnqueueNotificationTask(
		ctx,
		toNotificationPayload(payload),
		toAsynqOptions(opts)...)
}

func (a *taskQueueAdapter) EnqueueAcceptedRegistration(
	ctx context.Context,
	payload domain.AcceptedRegistrationFormTaskPayload,
	opts *domain.TaskEnqueueOptions,
) error {
	return a.client.EnqueueAcceptedRegistration(ctx, pkgasynq.AcceptedRegistrationFormPayload{
		ReferrerName:        payload.ReferrerName,
		ChildName:           payload.ChildName,
		ChildBSN:            payload.ChildBSN,
		AppointmentDate:     payload.AppointmentDate,
		AppointmentLocation: payload.AppointmentLocation,
		To:                  "",
	}, toAsynqOptions(opts)...)
}

func (a *taskQueueAdapter) EnqueueProcessRegistrationFormEmail(
	ctx context.Context,
	payload domain.ProcessRegistrationFormEmailTaskPayload,
	opts *domain.TaskEnqueueOptions,
) error {
	return a.client.EnqueueProcessRegistrationFormEmail(
		ctx,
		pkgasynq.ProcessRegistrationFormEmailPayload{
			ReferrerName: payload.ReferrerName,
			ClientName:   payload.ClientName,
			Location:     payload.Location,
			Link:         payload.Link,
			To:           payload.To,
		},
		toAsynqOptions(opts)...)
}

func (a *taskQueueAdapter) Close() error {
	return a.client.Close()
}

func toAsynqOptions(opts *domain.TaskEnqueueOptions) []hibikenasynq.Option {
	if opts == nil {
		return nil
	}

	result := make([]hibikenasynq.Option, 0, 2)
	if opts.Queue != "" {
		result = append(result, hibikenasynq.Queue(opts.Queue))
	}
	if opts.MaxRetry > 0 {
		result = append(result, hibikenasynq.MaxRetry(opts.MaxRetry))
	}
	return result
}

func toIncidentPayload(payload domain.IncidentTaskPayload) pkgasynq.IncidentPayload {
	return pkgasynq.IncidentPayload{
		ID:                      payload.ID,
		EmployeeID:              payload.EmployeeID,
		EmployeeFirstName:       payload.EmployeeFirstName,
		EmployeeLastName:        payload.EmployeeLastName,
		LocationID:              payload.LocationID,
		ReporterInvolvement:     payload.ReporterInvolvement,
		InformedParties:         payload.InformedParties,
		OccurredAt:              payload.OccurredAt,
		IncidentType:            payload.IncidentType,
		SeverityOfIncident:      payload.SeverityOfIncident,
		IncidentExplanation:     payload.IncidentExplanation,
		RecurrenceRisk:          payload.RecurrenceRisk,
		IncidentPreventSteps:    payload.IncidentPreventSteps,
		IncidentTakenMeasures:   payload.IncidentTakenMeasures,
		CauseCategories:         payload.CauseCategories,
		CauseExplanation:        payload.CauseExplanation,
		PhysicalInjury:          payload.PhysicalInjury,
		PhysicalInjuryDesc:      payload.PhysicalInjuryDesc,
		PsychologicalDamage:     payload.PsychologicalDamage,
		PsychologicalDamageDesc: payload.PsychologicalDamageDesc,
		NeededConsultation:      payload.NeededConsultation,
		FollowUpActions:         payload.FollowUpActions,
		FollowUpNotes:           payload.FollowUpNotes,
		IsEmployeeAbsent:        payload.IsEmployeeAbsent,
		AdditionalDetails:       payload.AdditionalDetails,
		ClientID:                payload.ClientID,
		LocationName:            payload.LocationName,
		Emails:                  payload.Emails,
	}
}

func toEmailDeliveryPayload(payload domain.EmailDeliveryTaskPayload) pkgasynq.EmailDeliveryPayload {
	return pkgasynq.EmailDeliveryPayload{
		To:           payload.To,
		Name:         payload.Name,
		UserEmail:    payload.UserEmail,
		UserPassword: payload.UserPassword,
	}
}

func toNotificationPayload(payload domain.NotificationTaskPayload) pkgasynq.NotificationPayload {
	return pkgasynq.NotificationPayload{
		RecipientUserIDs: payload.RecipientUserIDs,
		Type:             payload.Type,
		Data: pkgasynq.NotificationData{
			NewAppointment:      toNewAppointmentData(payload.Data.NewAppointment),
			NewClientAssignment: toNewClientAssignmentData(payload.Data.NewClientAssignment),
			ClientContractReminder: toClientContractReminderData(
				payload.Data.ClientContractReminder,
			),
			NewIncidentReport: toNewIncidentReportData(payload.Data.NewIncidentReport),
			NewScheduleNotification: toNewScheduleNotificationData(
				payload.Data.NewScheduleNotification,
			),
		},
		CreatedAt: payload.CreatedAt,
		Message:   payload.Message,
	}
}

func toNewAppointmentData(data *domain.NewAppointmentTaskData) *pkgasynq.NewAppointmentData {
	if data == nil {
		return nil
	}
	return &pkgasynq.NewAppointmentData{
		AppointmentID: data.AppointmentID,
		CreatedBy:     data.CreatedBy,
		StartTime:     data.StartTime,
		EndTime:       data.EndTime,
		Location:      data.Location,
	}
}

func toNewClientAssignmentData(
	data *domain.NewClientAssignmentTaskData,
) *pkgasynq.NewClientAssignmentData {
	if data == nil {
		return nil
	}
	return &pkgasynq.NewClientAssignmentData{
		ClientID:        data.ClientID,
		ClientFirstName: data.ClientFirstName,
		ClientLastName:  data.ClientLastName,
		ClientLocation:  data.ClientLocation,
	}
}

func toClientContractReminderData(
	data *domain.ClientContractReminderTaskData,
) *pkgasynq.ClientContractReminderData {
	if data == nil {
		return nil
	}
	return &pkgasynq.ClientContractReminderData{
		ClientID:           data.ClientID,
		ClientFirstName:    data.ClientFirstName,
		ClientLastName:     data.ClientLastName,
		ContractID:         data.ContractID,
		CareType:           data.CareType,
		ContractStart:      data.ContractStart,
		ContractEnd:        data.ContractEnd,
		ReminderType:       data.ReminderType,
		LastReminderSentAt: data.LastReminderSentAt,
	}
}

func toNewIncidentReportData(
	data *domain.NewIncidentReportTaskData,
) *pkgasynq.NewIncidentReportData {
	if data == nil {
		return nil
	}
	return &pkgasynq.NewIncidentReportData{
		ID:                 data.ID,
		EmployeeID:         data.EmployeeID,
		EmployeeFirstName:  data.EmployeeFirstName,
		EmployeeLastName:   data.EmployeeLastName,
		LocationID:         data.LocationID,
		LocationName:       data.LocationName,
		ClientID:           data.ClientID,
		ClientFirstName:    data.ClientFirstName,
		ClientLastName:     data.ClientLastName,
		SeverityOfIncident: data.SeverityOfIncident,
	}
}

func toNewScheduleNotificationData(
	data *domain.NewScheduleNotificationTaskData,
) *pkgasynq.NewScheduleNotificationData {
	if data == nil {
		return nil
	}
	return &pkgasynq.NewScheduleNotificationData{
		ScheduleID: data.ScheduleID,
		CreatedBy:  data.CreatedBy,
		StartTime:  data.StartTime,
		EndTime:    data.EndTime,
		Location:   data.Location,
	}
}
